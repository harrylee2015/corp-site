package handler

import (
	"encoding/json"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"corp-site/internal/config"
	"corp-site/internal/data"
	"corp-site/internal/database"
	"corp-site/internal/identity"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
)

func LoadCategoriesForIdentity(userIdentity string) []CategoryNavItem {
	allowed := identity.AllowedParents(userIdentity)
	if len(allowed) == 0 {
		return nil
	}
	all := LoadCategoryNav()
	var result []CategoryNavItem
	for _, item := range all {
		for _, name := range allowed {
			if item.Name == name {
				result = append(result, item)
				break
			}
		}
	}
	return result
}

func parseJSONUintSlice(s string) []uint {
	if s == "" {
		return nil
	}
	var ids []uint
	_ = json.Unmarshal([]byte(s), &ids)
	return ids
}

func parseJSONStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	var ss []string
	_ = json.Unmarshal([]byte(s), &ss)
	return ss
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func categoryAllowedForUser(user *model.User, categoryID uint) bool {
	cats := LoadCategoriesForIdentity(user.Identity)
	for _, p := range cats {
		if len(p.Children) == 0 && p.ID == categoryID {
			return true
		}
		for _, ch := range p.Children {
			if ch.ID == categoryID {
				return true
			}
		}
	}
	return false
}

func userCenterBase(c *gin.Context, user *model.User) gin.H {
	return gin.H{
		"user":          user,
		"identityLabel": identity.Label(user.Identity),
		"navCategories": LoadCategoriesForIdentity(user.Identity),
		"provinces":     data.Provinces,
		"csrf_token":    c.GetString("csrf_token"),
	}
}

func UserCenterHome(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	db := database.DB()
	var productCount, approvedCount, pendingCount int64
	db.Model(&model.Product{}).Where("user_id = ?", user.ID).Count(&productCount)
	db.Model(&model.Product{}).Where("user_id = ? AND status = ?", user.ID, "approved").Count(&approvedCount)
	db.Model(&model.Product{}).Where("user_id = ? AND status = ?", user.ID, "pending").Count(&pendingCount)

	var shop model.Shop
	hasShop := db.Where("user_id = ?", user.ID).First(&shop).Error == nil

	data := userCenterBase(c, user)
	data["ActiveNav"] = "home"
	data["productCount"] = productCount
	data["approvedCount"] = approvedCount
	data["pendingCount"] = pendingCount
	data["hasShop"] = hasShop
	data["verifyStatus"] = user.VerifyStatus
	renderUserPage(c, "用户首页", "center-home-content", data)
}

func UserShopPage(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	db := database.DB()
	var shop model.Shop
	db.Where("user_id = ?", user.ID).FirstOrInit(&shop, model.Shop{UserID: user.ID})

	data := userCenterBase(c, user)
	data["ActiveNav"] = "product"
	data["ActiveSub"] = "shop"
	data["shop"] = shop
	data["selectedRegions"] = parseJSONStringSlice(shop.Regions)
	data["selectedCategoryIDs"] = parseJSONUintSlice(shop.CategoryIDs)
	data["saved"] = c.Query("saved") == "1"
	if data["selectedRegions"] == nil {
		data["selectedRegions"] = []string{}
	}
	if data["selectedCategoryIDs"] == nil {
		data["selectedCategoryIDs"] = []uint{}
	}
	renderUserPage(c, "店铺信息", "shop-content", data)
}

func SaveShop(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.Redirect(302, "/login")
			return
		}
		regions := c.PostFormArray("regions")
		catIDs := c.PostFormArray("category_ids")
		var validCatIDs []uint
		for _, s := range catIDs {
			id, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				continue
			}
			if categoryAllowedForUser(user, uint(id)) {
				validCatIDs = append(validCatIDs, uint(id))
			}
		}

		shop := model.Shop{
			UserID:      user.ID,
			ShopName:    strings.TrimSpace(c.PostForm("shop_name")),
			Regions:     toJSON(regions),
			CategoryIDs: toJSON(validCatIDs),
			Contact:     strings.TrimSpace(c.PostForm("contact")),
			Phone:       strings.TrimSpace(c.PostForm("phone")),
			Tel:         strings.TrimSpace(c.PostForm("tel")),
			Address:     strings.TrimSpace(c.PostForm("address")),
			Intro:       strings.TrimSpace(c.PostForm("intro")),
			BannerPath:  strings.TrimSpace(c.PostForm("banner_path")),
		}

		db := database.DB()
		var existing model.Shop
		if err := db.Where("user_id = ?", user.ID).First(&existing).Error; err == nil {
			shop.ID = existing.ID
			shop.CreatedAt = existing.CreatedAt
			db.Save(&shop)
		} else {
			db.Create(&shop)
		}
		c.Redirect(302, "/my/shop?saved=1")
	}
}

func UserProductNew(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	if user.VerifyStatus != "approved" {
		data := userCenterBase(c, user)
		data["ActiveNav"] = "product"
		data["ActiveSub"] = "new"
		data["error"] = "请先完成企业认证后再发布产品"
		renderUserPage(c, "添加产品", "product-create-content", data)
		return
	}
	data := userCenterBase(c, user)
	data["ActiveNav"] = "product"
	data["ActiveSub"] = "new"
	renderUserPage(c, "添加产品", "product-create-content", data)
}

func CreateProduct(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	if user.VerifyStatus != "approved" {
		c.Redirect(302, "/my/products/new?error=verify")
		return
	}

	categoryID, _ := strconv.ParseUint(c.PostForm("category_id"), 10, 64)
	if !categoryAllowedForUser(user, uint(categoryID)) {
		data := userCenterBase(c, user)
		data["ActiveNav"] = "product"
		data["ActiveSub"] = "new"
		data["error"] = "所选分类与您的身份不匹配"
		renderUserPage(c, "添加产品", "product-create-content", data)
		return
	}

	amountWan, _ := strconv.ParseFloat(c.PostForm("amount_wan"), 64)
	ratePercent, _ := strconv.ParseFloat(c.PostForm("rate_percent"), 64)
	periodCount, _ := strconv.Atoi(c.PostForm("period_count"))
	regions := c.PostFormArray("regions")

	if amountWan <= 0 || ratePercent <= 0 || periodCount <= 0 {
		data := userCenterBase(c, user)
		data["ActiveNav"] = "product"
		data["ActiveSub"] = "new"
		data["error"] = "请填写完整的产品信息（额度、利率、期数）"
		renderUserPage(c, "添加产品", "product-create-content", data)
		return
	}

	product := model.Product{
		UserID:      user.ID,
		CategoryID:  uint(categoryID),
		Name:        strings.TrimSpace(c.PostForm("name")),
		ImagePath:   strings.TrimSpace(c.PostForm("image_path")),
		AmountWan:   amountWan,
		RateType:    c.PostForm("rate_type"),
		RatePercent: ratePercent,
		PeriodCount: periodCount,
		PeriodUnit:  "month",
		RepayMethod: c.PostForm("repay_method"),
		Regions:     toJSON(regions),
		Intro:       strings.TrimSpace(c.PostForm("intro")),
		Status:      "pending",
	}
	if product.Name == "" {
		data := userCenterBase(c, user)
		data["ActiveNav"] = "product"
		data["ActiveSub"] = "new"
		data["error"] = "请填写产品名称"
		renderUserPage(c, "添加产品", "product-create-content", data)
		return
	}
	if product.RateType != "daily" && product.RateType != "yearly" {
		product.RateType = "yearly"
	}
	if product.RepayMethod != "equal_installment" && product.RepayMethod != "equal_principal" {
		product.RepayMethod = "equal_installment"
	}

	database.DB().Create(&product)
	c.Redirect(302, "/my/products?created=1")
}

func UserProductList(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 10
	status := c.Query("status")

	query := database.DB().Model(&model.Product{}).Where("user_id = ?", user.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)

	var products []model.Product
	query.Preload("Category.Parent").Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&products)

	data := userCenterBase(c, user)
	data["ActiveNav"] = "product"
	data["ActiveSub"] = "list"
	data["products"] = products
	data["page"] = page
	data["totalPages"] = int(math.Ceil(float64(total) / float64(pageSize)))
	data["status"] = status
	data["created"] = c.Query("created") == "1"
	renderUserPage(c, "产品列表", "product-list-content", data)
}

func DeleteProduct(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}
	db := database.DB()
	var product model.Product
	if err := db.Where("user_id = ? AND id = ?", user.ID, c.Param("id")).First(&product).Error; err != nil {
		c.JSON(404, gin.H{"error": "产品不存在"})
		return
	}
	if product.Status != "pending" && product.Status != "rejected" {
		c.JSON(400, gin.H{"error": "仅待审核或已驳回的产品可删除"})
		return
	}
	db.Delete(&product)
	c.Redirect(302, "/my/products")
}

func UserProfilePage(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	data := userCenterBase(c, user)
	data["ActiveNav"] = "profile"
	data["saved"] = c.Query("saved") == "1"
	data["pwdChanged"] = c.Query("pwd") == "1"
	switch c.Query("error") {
	case "doc":
		data["error"] = "请先上传企业证明照片"
	case "img":
		data["error"] = "仅支持上传图片（jpg、png、gif、webp）"
	}
	renderUserPage(c, "基本信息", "profile-content", data)
}

func ChangeUserPassword(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.Redirect(302, "/login")
			return
		}
		oldPwd := c.PostForm("old_password")
		newPwd := c.PostForm("new_password")
		confirmPwd := c.PostForm("new_password_confirm")

		data := userCenterBase(c, user)
		data["ActiveNav"] = "profile"

		if !user.CheckPassword(oldPwd) {
			data["error"] = "原密码不正确"
			renderUserPage(c, "基本信息", "profile-content", data)
			return
		}
		if newPwd != confirmPwd {
			data["error"] = "两次新密码输入不一致"
			renderUserPage(c, "基本信息", "profile-content", data)
			return
		}
		if len(newPwd) < 8 || len(newPwd) > 20 {
			data["error"] = "新密码需8-20位"
			renderUserPage(c, "基本信息", "profile-content", data)
			return
		}
		user.SetPassword(newPwd)
		database.DB().Model(user).Update("password_hash", user.PasswordHash)
		c.Redirect(302, "/my/profile?pwd=1")
	}
}

func UploadVerifyDoc(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.JSON(401, gin.H{"error": "请先登录"})
			return
		}
		path := strings.TrimSpace(c.PostForm("doc_path"))
		if path == "" {
			c.Redirect(302, "/my/profile?error=doc")
			return
		}
		ext := strings.ToLower(filepath.Ext(path))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		default:
			c.Redirect(302, "/my/profile?error=img")
			return
		}
		database.DB().Model(user).Updates(map[string]interface{}{
			"verify_doc_path": path,
			"verify_status":   "approved",
		})
		c.Redirect(302, "/my/profile?saved=1")
	}
}

func containsUint(list []uint, v uint) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

func containsString(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// for templates
func IsSelectedRegion(regions []string, name string) bool { return containsString(regions, name) }
func IsSelectedCategory(ids []uint, id uint) bool         { return containsUint(ids, id) }
