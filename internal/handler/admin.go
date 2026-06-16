package handler

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"corp-site/internal/config"
	"corp-site/internal/data"
	"corp-site/internal/database"
	"corp-site/internal/identity"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AdminDashboard(c *gin.Context) {
	db := database.DB()
	var userCount, todayCount int64
	var productCount, productPending int64
	db.Model(&model.User{}).Where("role = ?", "user").Count(&userCount)
	db.Model(&model.Project{}).Where("created_at >= ?", time.Now().Truncate(24*time.Hour)).Count(&todayCount)
	db.Model(&model.Project{}).Count(&productCount)
	db.Model(&model.Project{}).Where("status = ?", "pending").Count(&productPending)

	renderPage(c, "layout/admin.html", "管理后台", "admindash-content", gin.H{
		"userCount":      userCount,
		"todayCount":     todayCount,
		"productCount":   productCount,
		"projectCount":   productCount,
		"productPending": productPending,
		"projectPending": productPending,
		"categoryStats":  LoadCategoryStats(),
		"csrf_token":     c.GetString("csrf_token"),
	})
}

func AdminReview(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var total int64
	db.Model(&model.Post{}).Where("status = ?", "pending").Count(&total)

	var posts []model.Post
	db.Where("status = ?", "pending").Preload("Category").Preload("User").
		Preload("Attachments").Order("created_at ASC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/admin.html", "审核管理", "adminreview-content", gin.H{
		"posts":      posts,
		"page":       page,
		"totalPages": totalPages,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func AdminPosts(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	status := c.Query("status")
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Model(&model.Post{}).Preload("Category").Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		for _, word := range strings.Fields(keyword) {
			like := "%" + word + "%"
			query = query.Where("(title ILIKE ? OR content ILIKE ? OR contact ILIKE ?)", like, like, like)
		}
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	var total int64
	query.Count(&total)

	var posts []model.Post
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/admin.html", "信息管理", "adminposts-content", gin.H{
		"posts":         posts,
		"navCategories": LoadCategoryNav(),
		"page":       page,
		"totalPages": totalPages,
		"status":     status,
		"categoryID": categoryID,
		"keyword":    keyword,
		"startDate":  startDate,
		"endDate":    endDate,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func AdminProjects(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	status := c.Query("status")
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.Model(&model.Project{}).Preload("Category.Parent").Preload("User")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		for _, word := range strings.Fields(keyword) {
			like := "%" + word + "%"
			query = query.Where("(name ILIKE ? OR intro ILIKE ?)", like, like)
		}
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	var total int64
	query.Count(&total)

	var projects []model.Project
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/admin.html", "项目管理", "adminprojects-content", gin.H{
		"projects":      projects,
		"navCategories": LoadCategoryNav(),
		"page":          page,
		"totalPages":    totalPages,
		"status":        status,
		"categoryID":    categoryID,
		"keyword":       keyword,
		"startDate":     startDate,
		"endDate":       endDate,
		"saved":         c.Query("saved"),
		"csrf_token":    c.GetString("csrf_token"),
	})
}

func AdminProjectEditPage(c *gin.Context) {
	db := database.DB()
	var project model.Project
	if err := db.Preload("Category.Parent").Preload("User").
		First(&project, "id = ?", c.Param("id")).Error; err != nil {
		c.Redirect(302, "/admin/projects")
		return
	}

	pageData := gin.H{
		"project":       project,
		"navCategories": LoadCategoryNav(),
		"provinces":     data.Provinces,
		"citiesJSON":    data.CitiesJSON(),
		"isFunder":      project.User.Identity == identity.Funder,
		"identityLabel": identity.Label(project.User.Identity),
		"csrf_token":    c.GetString("csrf_token"),
	}
	bindProjectRegionFields(pageData, project)

	renderPage(c, "layout/admin.html", "编辑项目", "adminprojectedit-content", pageData)
}

func AdminUpdateProject(c *gin.Context) {
	db := database.DB()
	var project model.Project
	if err := db.Preload("User").First(&project, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "项目不存在"})
		return
	}

	categoryID, _ := strconv.ParseUint(c.PostForm("category_id"), 10, 64)
	if categoryID == 0 {
		renderAdminProjectEditError(c, project, "请选择分类")
		return
	}

	regionsJSON, err := data.BuildRegionsJSON(
		project.User.Identity,
		c.PostFormArray("regions"),
		c.PostForm("province"),
		c.PostForm("city"),
	)
	if err != nil {
		renderAdminProjectEditError(c, project, err.Error())
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		renderAdminProjectEditError(c, project, "请填写项目名称")
		return
	}

	contactPerson := strings.TrimSpace(c.PostForm("contact_person"))
	contactPhone := strings.TrimSpace(c.PostForm("contact_phone"))
	if contactPerson == "" || contactPhone == "" {
		renderAdminProjectEditError(c, project, "请填写联系人和联系电话")
		return
	}

	status := c.PostForm("status")
	if status != "pending" && status != "approved" && status != "rejected" && status != "delisted" {
		renderAdminProjectEditError(c, project, "无效的状态")
		return
	}

	updates := map[string]interface{}{
		"category_id":    uint(categoryID),
		"name":           name,
		"regions":        regionsJSON,
		"intro":          strings.TrimSpace(c.PostForm("intro")),
		"contact_person": contactPerson,
		"contact_phone":  contactPhone,
		"status":         status,
	}

	imagePath := strings.TrimSpace(c.PostForm("image_path"))
	if imagePath != "" {
		updates["image_path"] = imagePath
	}

	budgetStr := strings.TrimSpace(c.PostForm("budget_amount_wan"))
	if budgetStr != "" {
		budget, _ := strconv.ParseFloat(budgetStr, 64)
		if budget > 0 {
			updates["budget_amount_wan"] = budget
		}
	}

	if project.User.Identity == identity.Funder {
		amountWan, _ := strconv.ParseFloat(c.PostForm("amount_wan"), 64)
		ratePercent, _ := strconv.ParseFloat(c.PostForm("rate_percent"), 64)
		periodCount, _ := strconv.Atoi(c.PostForm("period_count"))
		rateType := c.PostForm("rate_type")
		repayMethod := c.PostForm("repay_method")
		if amountWan <= 0 || ratePercent <= 0 || periodCount <= 0 {
			renderAdminProjectEditError(c, project, "请填写完整的金融信息（额度、利率、期数）")
			return
		}
		if rateType != "daily" && rateType != "yearly" {
			rateType = "yearly"
		}
		if repayMethod != "equal_installment" && repayMethod != "equal_principal" {
			repayMethod = "equal_installment"
		}
		updates["amount_wan"] = amountWan
		updates["rate_percent"] = ratePercent
		updates["period_count"] = periodCount
		updates["rate_type"] = rateType
		updates["repay_method"] = repayMethod
	}

	if err := db.Model(&project).Updates(updates).Error; err != nil {
		renderAdminProjectEditError(c, project, "保存失败")
		return
	}
	c.Redirect(302, "/admin/projects?saved=1")
}

func renderAdminProjectEditError(c *gin.Context, project model.Project, msg string) {
	database.DB().Preload("Category.Parent").Preload("User").
		First(&project, "id = ?", project.ID)
	pageData := gin.H{
		"project":       project,
		"navCategories": LoadCategoryNav(),
		"provinces":     data.Provinces,
		"citiesJSON":    data.CitiesJSON(),
		"isFunder":      project.User.Identity == identity.Funder,
		"identityLabel": identity.Label(project.User.Identity),
		"error":         msg,
		"csrf_token":    c.GetString("csrf_token"),
	}
	bindProjectRegionFields(pageData, project)
	renderPage(c, "layout/admin.html", "编辑项目", "adminprojectedit-content", pageData)
}

func AdminDeleteProject(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.DB()
		var project model.Project
		if err := db.First(&project, "id = ?", c.Param("id")).Error; err != nil {
			c.JSON(404, gin.H{"error": "项目不存在"})
			return
		}

		if project.ImagePath != "" {
			os.Remove(filepath.Join(cfg.Upload.Path, project.ImagePath))
		}

		if err := db.Delete(&project).Error; err != nil {
			c.JSON(500, gin.H{"error": "删除失败"})
			return
		}
		c.JSON(200, gin.H{"message": "删除成功"})
	}
}

func AdminExportPage(c *gin.Context) {
	renderPage(c, "layout/admin.html", "导出报表", "adminexport-content", gin.H{
		"navCategories": LoadCategoryNav(),
		"csrf_token":    c.GetString("csrf_token"),
	})
}

func AdminUsers(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	keyword := c.Query("keyword")

	query := db.Model(&model.User{}).Where("role = ?", "user")
	if keyword != "" {
		query = query.Where("phone ILIKE ? OR nickname ILIKE ? OR real_name ILIKE ? OR company ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	var total int64
	query.Count(&total)

	var users []model.User
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/admin.html", "用户管理", "adminusers-content", gin.H{
		"users":      users,
		"page":       page,
		"totalPages": totalPages,
		"keyword":    keyword,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func ReviewPost(c *gin.Context) {
	var req struct {
		Action string `json:"action" form:"action" binding:"required"`
		Reason string `json:"reason" form:"reason"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	userIDStr, _ := c.Get("user_id")
	reviewerID, _ := uuid.Parse(userIDStr.(string))
	now := time.Now()

	db := database.DB()

	action := req.Action
	if action != "approve" && action != "reject" {
		c.JSON(400, gin.H{"error": "无效的审核操作"})
		return
	}

	updates := map[string]interface{}{
		"reviewed_by": reviewerID,
		"reviewed_at": now,
	}
	if action == "approve" {
		updates["status"] = "approved"
	} else {
		updates["status"] = "rejected"
		updates["reject_reason"] = req.Reason
	}

	result := db.Model(&model.Post{}).
		Where("id = ? AND status = ?", c.Param("id"), "pending").
		Updates(updates)

	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "信息不存在或已被审核"})
		return
	} else if result.Error != nil {
		c.JSON(500, gin.H{"error": "审核失败"})
		return
	}

	c.JSON(200, gin.H{"message": "审核完成"})
}

func AdminProjectReview(c *gin.Context) {
	db := database.DB()
	var projects []model.Project
	db.Where("status = ?", "pending").Preload("Category.Parent").Preload("User").
		Order("created_at ASC").Limit(50).Find(&projects)
	renderPage(c, "layout/admin.html", "项目审核", "adminprojectreview-content", gin.H{
		"projects":   projects,
		"products":   projects,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func AdminProductReview(c *gin.Context) {
	c.Redirect(302, "/admin/project-review")
}

func ReviewProject(c *gin.Context) {
	var req struct {
		Action string `json:"action" form:"action" binding:"required"`
		Reason string `json:"reason" form:"reason"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	userIDStr, _ := c.Get("user_id")
	reviewerID, _ := uuid.Parse(userIDStr.(string))
	now := time.Now()
	action := req.Action
	if action != "approve" && action != "reject" {
		c.JSON(400, gin.H{"error": "无效的审核操作"})
		return
	}
	updates := map[string]interface{}{"reviewed_by": reviewerID, "reviewed_at": now}
	if action == "approve" {
		updates["status"] = "approved"
	} else {
		updates["status"] = "rejected"
		updates["reject_reason"] = req.Reason
	}
	result := database.DB().Model(&model.Project{}).
		Where("id = ? AND status = ?", c.Param("id"), "pending").Updates(updates)
	if result.RowsAffected == 0 {
		c.JSON(404, gin.H{"error": "项目不存在或已被审核"})
		return
	}
	c.JSON(200, gin.H{"message": "审核完成"})
}

func ReviewProduct(c *gin.Context) { ReviewProject(c) }

func UpdateUserVerify(c *gin.Context) {
	var req struct {
		Status string `json:"status" form:"status" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}
	if req.Status != "approved" && req.Status != "rejected" && req.Status != "pending" {
		c.JSON(400, gin.H{"error": "无效状态"})
		return
	}
	database.DB().Model(&model.User{}).Where("id = ? AND role = ?", c.Param("id"), "user").
		Update("verify_status", req.Status)
	c.JSON(200, gin.H{"message": "更新成功"})
}

func UpdateUserStatus(c *gin.Context) {
	var req struct {
		Status string `json:"status" form:"status" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	if req.Status != "active" && req.Status != "disabled" {
		c.JSON(400, gin.H{"error": "无效的状态"})
		return
	}

	db := database.DB()
	db.Model(&model.User{}).Where("id = ? AND role = ?", c.Param("id"), "user").
		Update("status", req.Status)

	c.JSON(200, gin.H{"message": "更新成功"})
}

func AdminPostDetail(c *gin.Context) {
	db := database.DB()
	var post model.Post
	if err := db.Preload("Category").Preload("User").Preload("Attachments").
		First(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "信息不存在"})
		return
	}
	c.JSON(200, gin.H{"post": post})
}

func AdminDeletePost(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		db := database.DB()
		var post model.Post
		if err := db.First(&post, "id = ?", c.Param("id")).Error; err != nil {
			c.JSON(404, gin.H{"error": "信息不存在"})
			return
		}

		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(500, gin.H{"error": "系统错误"})
			return
		}

		var attachments []model.Attachment
		if err := tx.Where("post_id = ?", post.ID).Find(&attachments).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "系统错误"})
			return
		}
		for _, a := range attachments {
			os.Remove(filepath.Join(cfg.Upload.Path, a.FilePath))
			if err := tx.Delete(&a).Error; err != nil {
				tx.Rollback()
				c.JSON(500, gin.H{"error": "系统错误"})
				return
			}
		}

		if err := tx.Delete(&post).Error; err != nil {
			tx.Rollback()
			c.JSON(500, gin.H{"error": "系统错误"})
			return
		}

		tx.Commit()
		c.JSON(200, gin.H{"message": "删除成功"})
	}
}

func AdminPasswordPage(c *gin.Context) {
	renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
		"csrf_token": c.GetString("csrf_token"),
	})
}

func ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword        string `form:"old_password" binding:"required"`
		NewPassword        string `form:"new_password" binding:"required,min=8,max=20"`
		NewPasswordConfirm string `form:"new_password_confirm" binding:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
			"error":      "请填写完整信息",
			"csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	if req.NewPassword != req.NewPasswordConfirm {
		renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
			"error":      "两次密码输入不一致",
			"csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	hasLetter := false
	hasDigit := false
	for _, ch := range req.NewPassword {
		if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
			hasLetter = true
		}
		if ch >= '0' && ch <= '9' {
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
			"error":      "密码需包含字母和数字",
			"csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	user := authUserFromContext(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}

	if !user.CheckPassword(req.OldPassword) {
		renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
			"error":      "当前密码错误",
			"csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	if err := user.SetPassword(req.NewPassword); err != nil {
		renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
			"error":      "系统错误，请重试",
			"csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	db := database.DB()
	db.Model(user).Update("password_hash", user.PasswordHash)

	renderPage(c, "layout/admin.html", "修改密码", "adminpassword-content", gin.H{
		"flash":      "密码修改成功",
		"csrf_token": c.GetString("csrf_token"),
	})
}

func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + strings.Repeat("*", 4) + phone[7:]
}

func MaskName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	runes := []rune(name)
	if len(runes) == 1 {
		return string(runes[0]) + "*"
	}
	return string(runes[0]) + strings.Repeat("*", len(runes)-1)
}
