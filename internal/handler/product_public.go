package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"corp-site/internal/data"
	"corp-site/internal/database"
	"corp-site/internal/identity"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Index(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 12
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	query := buildPublicProjectQuery(db, categoryID, keyword)

	var total int64
	query.Count(&total)

	var projects []model.Project
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/base.html", "首页 - 金筹设备租赁", "index-content", gin.H{
		"projects":   projects,
		"products":   projects,
		"page":       page,
		"totalPages": totalPages,
		"categoryID": categoryID,
		"keyword":    keyword,
		"total":      total,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func ProjectDetail(c *gin.Context) {
	serveProjectDetail(c, c.Param("id"))
}

func ProductDetail(c *gin.Context) {
	serveProjectDetail(c, c.Param("id"))
}

func serveProjectDetail(c *gin.Context, id string) {
	db := database.DB()
	var project model.Project
	if err := db.Preload("Category.Parent").Preload("User").
		First(&project, "id = ?", id).Error; err != nil {
		renderPage(c, "layout/base.html", "项目不存在", "projectdetail-content", gin.H{
			"error": "项目不存在", "csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	visible := project.Status == "approved"
	userIDStr, _ := c.Get("user_id")
	role, _ := c.Get("role")
	isOwner := userIDStr != nil && userIDStr.(string) == project.UserID.String()
	isAdmin := role == "admin"
	if !visible && (isOwner || isAdmin) {
		visible = true
	}
	if !visible {
		renderPage(c, "layout/base.html", "项目不存在", "projectdetail-content", gin.H{
			"error": "项目不存在或已下架", "csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	var company model.Company
	hasCompany := db.Where("user_id = ?", project.UserID).First(&company).Error == nil

	regionDisplay := data.FormatRegionsDisplay(project.Regions, project.User.Identity)
	data := gin.H{
		"project":     project,
		"product":     project,
		"showPrivate": isOwner || isAdmin,
		"isFunder":    project.User.Identity == identity.Funder,
		"regions":     regionDisplay,
		"csrf_token":  c.GetString("csrf_token"),
	}
	if hasCompany && (isOwner || isAdmin) {
		data["company"] = company
		data["shop"] = company
		data["maskedShopContact"] = MaskName(company.Contact)
		data["maskedShopPhone"] = MaskPhone(company.Phone)
	}
	renderPage(c, "layout/base.html", project.Name, "projectdetail-content", data)
}

func ProjectList(c *gin.Context) {
	serveProjectListAPI(c)
}

func ProductList(c *gin.Context) {
	serveProjectListAPI(c)
}

func serveProjectListAPI(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 12
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	query := buildPublicProjectQuery(db, categoryID, keyword)

	var total int64
	query.Count(&total)

	var projects []model.Project
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects)

	hasMore := int64(page*pageSize) < total

	type row struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Intro    string `json:"intro"`
		Category string `json:"category"`
		Amount   string `json:"amount"`
		Rate     string `json:"rate"`
		Period   string `json:"period"`
		Budget   string `json:"budget"`
		Region   string `json:"region"`
		IsFunder bool   `json:"is_funder"`
		Nickname string `json:"nickname"`
		Date     string `json:"date"`
	}

	rows := make([]row, len(projects))
	for i, p := range projects {
		r := row{
			ID:       p.ID.String(),
			Name:     p.Name,
			Intro:    p.Intro,
			Category: formatPostCategory(p.Category),
			Region:   strings.Join(data.FormatRegionsDisplay(p.Regions, p.User.Identity), "、"),
			IsFunder: p.User.Identity == identity.Funder,
			Nickname: maskPublisherName(p.User),
			Date:     p.CreatedAt.Format("01-02"),
		}
		if p.BudgetAmountWan != nil && *p.BudgetAmountWan > 0 {
			r.Budget = fmt.Sprintf("%.0f万", *p.BudgetAmountWan)
		}
		if p.IsFunderProject() {
			r.Amount = fmt.Sprintf("%.0f万", *p.AmountWan)
			r.Rate = projectRateLabel(p)
			if p.PeriodCount != nil {
				r.Period = fmt.Sprintf("%d期", *p.PeriodCount)
			}
		}
		rows[i] = r
	}

	c.JSON(200, gin.H{"projects": rows, "products": rows, "has_more": hasMore})
}

func buildPublicProjectQuery(db *gorm.DB, categoryID, keyword string) *gorm.DB {
	query := db.Model(&model.Project{}).Where("status = ?", "approved").
		Preload("Category.Parent").Preload("User")
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		for _, word := range strings.Fields(keyword) {
			like := "%" + word + "%"
			query = query.Where("(name ILIKE ? OR intro ILIKE ?)", like, like)
		}
	}
	return query
}

func projectRateLabel(p model.Project) string {
	if p.RateType == nil || p.RatePercent == nil {
		return ""
	}
	return formatProductRate(*p.RateType, *p.RatePercent)
}

func formatProductRate(rateType string, percent float64) string {
	label := "年"
	if rateType == "daily" {
		label = "日"
	}
	return fmt.Sprintf("%s利率 %.2f%%", label, percent)
}

func formatRepayMethod(method string) string {
	switch method {
	case "equal_installment":
		return "等额本息"
	case "equal_principal":
		return "等额本金"
	default:
		return method
	}
}

func FormatProductRate(rateType string, percent float64) string {
	return formatProductRate(rateType, percent)
}

func FormatRepayMethod(method *string) string {
	if method == nil || *method == "" {
		return ""
	}
	return formatRepayMethod(*method)
}

func FormatProjectRate(p model.Project) string {
	return projectRateLabel(p)
}

func maskPublisherName(user model.User) string {
	if user.RealName != "" {
		return MaskName(user.RealName)
	}
	if user.Nickname != "" {
		return MaskName(user.Nickname)
	}
	return MaskPhone(user.Phone)
}
