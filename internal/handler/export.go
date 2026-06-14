package handler

import (
	"fmt"
	"strings"
	"time"

	"corp-site/internal/database"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ExportExcel(c *gin.Context) {
	var req struct {
		CategoryID string `form:"category_id"`
		Status     string `form:"status"`
		StartDate  string `form:"start_date"`
		EndDate    string `form:"end_date"`
		Keyword    string `form:"keyword"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	db := database.DB()
	query := buildExportQuery(db, req).Preload("Category").Preload("User")

	var posts []model.Post
	query.Order("created_at DESC").Find(&posts)

	f := excelize.NewFile()
	defer f.Close()

	// Sheet1: data
	sheet1 := "信息列表"
	f.SetSheetName("Sheet1", sheet1)

	headers := []string{"标题", "分类", "内容", "联系人", "联系电话", "发布人", "状态", "发布时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet1, cell, h)
	}

	statusMap := map[string]string{
		"pending":  "待审核",
		"approved": "已通过",
		"rejected": "已驳回",
	}

	for i, post := range posts {
		row := i + 2
		nickname := ""
		if post.User.Nickname != "" {
			nickname = post.User.Nickname
		} else {
			nickname = MaskPhone(post.User.Phone)
		}
		categoryName := post.Category.Name

		values := []string{
			post.Title,
			categoryName,
			post.Content,
			post.Contact,
			post.ContactPhone,
			nickname,
			statusMap[post.Status],
			post.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet1, cell, v)
		}
	}

	// Sheet2: summary
	sheet2 := "统计汇总"
	f.NewSheet(sheet2)

	type statRow struct {
		Label string
		Count int64
	}

	var stats []statRow

	var catStats []struct {
		CategoryName string
		Count        int64
	}
	db.Model(&model.Post{}).
		Select("categories.name as category_name, COUNT(*) as count").
		Joins("JOIN categories ON categories.id = posts.category_id").
		Where("posts.status = ?", "approved").
		Group("categories.name").Scan(&catStats)
	for _, cs := range catStats {
		stats = append(stats, statRow{Label: cs.CategoryName + " (已审核)", Count: cs.Count})
	}

	var statusStats []struct {
		Status string
		Count  int64
	}
	db.Model(&model.Post{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats)
	for _, ss := range statusStats {
		stats = append(stats, statRow{Label: statusMap[ss.Status], Count: ss.Count})
	}

	f.SetCellValue(sheet2, "A1", "统计项")
	f.SetCellValue(sheet2, "B1", "数量")
	for i, s := range stats {
		f.SetCellValue(sheet2, fmt.Sprintf("A%d", i+2), s.Label)
		f.SetCellValue(sheet2, fmt.Sprintf("B%d", i+2), s.Count)
	}

	f.SetActiveSheet(0)

	filename := fmt.Sprintf("导出_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Transfer-Encoding", "binary")

	if err := f.Write(c.Writer); err != nil {
		c.JSON(500, gin.H{"error": "生成Excel失败"})
	}
}

func ExportUsersExcel(c *gin.Context) {
	db := database.DB()
	keyword := c.Query("keyword")
	status := c.Query("status")

	query := db.Model(&model.User{}).Where("role = ?", "user")
	if keyword != "" {
		like := "%" + keyword + "%"
		query = query.Where("phone ILIKE ? OR real_name ILIKE ? OR company ILIKE ? OR nickname ILIKE ?",
			like, like, like, like)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var users []model.User
	query.Order("created_at DESC").Find(&users)

	shopMap := map[string]model.Shop{}
	var shops []model.Shop
	db.Find(&shops)
	for _, s := range shops {
		shopMap[s.UserID.String()] = s
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "用户列表"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"手机号", "真实姓名", "企业名称", "身份", "认证状态", "账号状态", "公司名称", "公司联系人", "公司手机", "注册时间"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	identityMap := map[string]string{"demander": "需求方", "supplier": "设备供应商", "funder": "资金方"}
	verifyMap := map[string]string{"none": "未认证", "approved": "已认证", "pending": "审核中", "rejected": "未通过"}
	statusMap := map[string]string{"active": "正常", "disabled": "已禁用"}

	for i, u := range users {
		row := i + 2
		shop := shopMap[u.ID.String()]
		values := []string{
			u.Phone,
			u.RealName,
			u.Company,
			identityMap[u.Identity],
			verifyMap[u.VerifyStatus],
			statusMap[u.Status],
			shop.ShopName,
			shop.Contact,
			shop.Phone,
			u.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for j, v := range values {
			cell, _ := excelize.CoordinatesToCellName(j+1, row)
			f.SetCellValue(sheet, cell, v)
		}
	}

	filename := fmt.Sprintf("用户导出_%s.xlsx", time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Header("Content-Transfer-Encoding", "binary")
	if err := f.Write(c.Writer); err != nil {
		c.JSON(500, gin.H{"error": "生成Excel失败"})
	}
}

func ExportPreview(c *gin.Context) {
	var req struct {
		CategoryID string `form:"category_id"`
		Status     string `form:"status"`
		StartDate  string `form:"start_date"`
		EndDate    string `form:"end_date"`
		Keyword    string `form:"keyword"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	db := database.DB()
	query := buildExportQuery(db, req)

	var total int64
	query.Count(&total)

	var posts []model.Post
	query.Preload("Category").Preload("User").Order("created_at DESC").Limit(20).Find(&posts)

	type previewRow struct {
		Title    string `json:"title"`
		Category string `json:"category"`
		Contact  string `json:"contact"`
		Phone    string `json:"phone"`
		Status   string `json:"status"`
		Created  string `json:"created"`
	}

	statusMap := map[string]string{
		"pending": "待审核", "approved": "已通过", "rejected": "已驳回",
	}

	rows := make([]previewRow, len(posts))
	for i, p := range posts {
		contact := p.Contact
		phone := MaskPhone(p.ContactPhone)
		if phone == "" {
			phone = ""
		}
		rows[i] = previewRow{
			Title:    p.Title,
			Category: p.Category.Name,
			Contact:  contact,
			Phone:    phone,
			Status:   statusMap[p.Status],
			Created:  p.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	c.JSON(200, gin.H{"total": total, "rows": rows})
}

func buildExportQuery(db *gorm.DB, req struct {
	CategoryID string `form:"category_id"`
	Status     string `form:"status"`
	StartDate  string `form:"start_date"`
	EndDate    string `form:"end_date"`
	Keyword    string `form:"keyword"`
}) *gorm.DB {
	query := db.Model(&model.Post{})
	if req.CategoryID != "" {
		ids := strings.Split(req.CategoryID, ",")
		query = query.Where("category_id IN ?", ids)
	}
	if req.Status != "" {
		statuses := strings.Split(req.Status, ",")
		query = query.Where("status IN ?", statuses)
	}
	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}
	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate+" 23:59:59")
	}
	if req.Keyword != "" {
		for _, word := range strings.Fields(req.Keyword) {
			like := "%" + word + "%"
			query = query.Where("(title ILIKE ? OR content ILIKE ? OR contact ILIKE ?)", like, like, like)
		}
	}
	return query
}
