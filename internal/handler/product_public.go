package handler

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"corp-site/internal/database"
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

	query := buildPublicProductQuery(db, categoryID, keyword)

	var total int64
	query.Count(&total)

	var products []model.Product
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&products)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/base.html", "首页 - 金筹设备租赁", "index-content", gin.H{
		"products":   products,
		"page":       page,
		"totalPages": totalPages,
		"categoryID": categoryID,
		"keyword":    keyword,
		"total":      total,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func ProductDetail(c *gin.Context) {
	db := database.DB()
	var product model.Product
	if err := db.Preload("Category.Parent").Preload("User").
		First(&product, "id = ?", c.Param("id")).Error; err != nil {
		renderPage(c, "layout/base.html", "产品不存在", "productdetail-content", gin.H{
			"error": "产品不存在", "csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	visible := product.Status == "approved"
	userIDStr, _ := c.Get("user_id")
	role, _ := c.Get("role")
	isOwner := userIDStr != nil && userIDStr.(string) == product.UserID.String()
	isAdmin := role == "admin"
	if !visible && (isOwner || isAdmin) {
		visible = true
	}
	if !visible {
		renderPage(c, "layout/base.html", "产品不存在", "productdetail-content", gin.H{
			"error": "产品不存在或已下架", "csrf_token": c.GetString("csrf_token"),
		})
		return
	}

	var shop model.Shop
	hasShop := db.Where("user_id = ?", product.UserID).First(&shop).Error == nil

	data := gin.H{
		"product":     product,
		"showPrivate": isOwner || isAdmin,
		"regions":     parseJSONStringSlice(product.Regions),
		"csrf_token":  c.GetString("csrf_token"),
	}
	if hasShop {
		data["shop"] = shop
		data["maskedShopContact"] = MaskName(shop.Contact)
		data["maskedShopPhone"] = MaskPhone(shop.Phone)
	}
	renderPage(c, "layout/base.html", product.Name, "productdetail-content", data)
}

func ProductList(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 12
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	query := buildPublicProductQuery(db, categoryID, keyword)

	var total int64
	query.Count(&total)

	var products []model.Product
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&products)

	hasMore := int64(page*pageSize) < total

	type row struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Intro    string `json:"intro"`
		Category string `json:"category"`
		Amount   string `json:"amount"`
		Rate     string `json:"rate"`
		Period   string `json:"period"`
		Nickname string `json:"nickname"`
		Date     string `json:"date"`
	}

	rows := make([]row, len(products))
	for i, p := range products {
		rows[i] = row{
			ID:       p.ID.String(),
			Name:     p.Name,
			Intro:    p.Intro,
			Category: formatPostCategory(p.Category),
			Amount:   fmt.Sprintf("%.0f万", p.AmountWan),
			Rate:     formatProductRate(p.RateType, p.RatePercent),
			Period:   fmt.Sprintf("%d期", p.PeriodCount),
			Nickname: maskPublisherName(p.User),
			Date:     p.CreatedAt.Format("01-02"),
		}
	}

	c.JSON(200, gin.H{"products": rows, "has_more": hasMore})
}

func buildPublicProductQuery(db *gorm.DB, categoryID, keyword string) *gorm.DB {
	query := db.Model(&model.Product{}).Where("status = ?", "approved").
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

func FormatRepayMethod(method string) string {
	return formatRepayMethod(method)
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
