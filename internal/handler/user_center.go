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

func enrichUserCenterData(h gin.H, user *model.User) {
	h["identity"] = user.Identity
	h["regionMode"] = data.RegionMode(user.Identity)
	h["isFunder"] = user.Identity == identity.Funder
	h["citiesJSON"] = data.CitiesJSON()
}

func bindCompanyRegionFields(h gin.H, company model.Company, user *model.User) {
	if data.IsMultiProvince(user.Identity) {
		h["selectedRegions"] = data.ParseProvinces(company.Regions)
		if h["selectedRegions"] == nil {
			h["selectedRegions"] = []string{}
		}
	} else {
		p, c := data.ParseSingleCity(company.Regions)
		h["selectedProvince"] = p
		h["selectedCity"] = c
	}
}

func userCenterBase(c *gin.Context, user *model.User) gin.H {
	h := gin.H{
		"user":          user,
		"identityLabel": identity.Label(user.Identity),
		"navCategories": LoadCategoriesForIdentity(user.Identity),
		"provinces":     data.Provinces,
		"csrf_token":    c.GetString("csrf_token"),
	}
	enrichUserCenterData(h, user)
	return h
}

func UserCenterHome(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	db := database.DB()
	var projectCount, approvedCount, pendingCount int64
	db.Model(&model.Project{}).Where("user_id = ?", user.ID).Count(&projectCount)
	db.Model(&model.Project{}).Where("user_id = ? AND status = ?", user.ID, "approved").Count(&approvedCount)
	db.Model(&model.Project{}).Where("user_id = ? AND status = ?", user.ID, "pending").Count(&pendingCount)

	var company model.Company
	hasCompany := db.Where("user_id = ?", user.ID).First(&company).Error == nil

	data := userCenterBase(c, user)
	data["ActiveNav"] = "home"
	data["projectCount"] = projectCount
	data["productCount"] = projectCount
	data["approvedCount"] = approvedCount
	data["pendingCount"] = pendingCount
	data["hasCompany"] = hasCompany
	data["hasShop"] = hasCompany
	data["verifyStatus"] = user.VerifyStatus
	renderUserPage(c, "用户首页", "center-home-content", data)
}

func UserCompanyPage(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	db := database.DB()
	var company model.Company
	db.Where("user_id = ?", user.ID).FirstOrInit(&company, model.Company{UserID: user.ID})
	if company.ShopName == "" && user.Company != "" {
		company.ShopName = user.Company
	}

	data := userCenterBase(c, user)
	data["ActiveNav"] = "project"
	data["ActiveSub"] = "company"
	data["company"] = company
	data["shop"] = company
	data["selectedCategoryIDs"] = parseJSONUintSlice(company.CategoryIDs)
	data["saved"] = c.Query("saved") == "1"
	bindCompanyRegionFields(data, company, user)
	if data["selectedCategoryIDs"] == nil {
		data["selectedCategoryIDs"] = []uint{}
	}
	renderUserPage(c, "公司信息", "company-content", data)
}

func UserShopPage(c *gin.Context) { c.Redirect(302, "/my/company") }

func SaveCompany(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.Redirect(302, "/login")
			return
		}
		regionsJSON, err := data.BuildRegionsJSON(
			user.Identity,
			c.PostFormArray("regions"),
			c.PostForm("province"),
			c.PostForm("city"),
		)
		if err != nil {
			data := userCenterBase(c, user)
			data["ActiveNav"] = "project"
			data["ActiveSub"] = "company"
			data["error"] = err.Error()
			var company model.Company
			database.DB().Where("user_id = ?", user.ID).FirstOrInit(&company, model.Company{UserID: user.ID})
			data["company"] = company
			data["shop"] = company
			bindCompanyRegionFields(data, company, user)
			renderUserPage(c, "公司信息", "company-content", data)
			return
		}

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

		company := model.Company{
			UserID:        user.ID,
			ShopName:      strings.TrimSpace(c.PostForm("shop_name")),
			EstablishedAt: strings.TrimSpace(c.PostForm("established_at")),
			Regions:       regionsJSON,
			CategoryIDs:   toJSON(validCatIDs),
			Contact:       strings.TrimSpace(c.PostForm("contact")),
			Phone:         strings.TrimSpace(c.PostForm("phone")),
			Tel:           strings.TrimSpace(c.PostForm("tel")),
			Address:       strings.TrimSpace(c.PostForm("address")),
			Intro:         strings.TrimSpace(c.PostForm("intro")),
			BannerPath:    strings.TrimSpace(c.PostForm("banner_path")),
		}

		db := database.DB()
		var existing model.Company
		if err := db.Where("user_id = ?", user.ID).First(&existing).Error; err == nil {
			company.ID = existing.ID
			company.CreatedAt = existing.CreatedAt
			db.Save(&company)
		} else {
			db.Create(&company)
		}
		c.Redirect(302, "/my/company?saved=1")
	}
}

func SaveShop(cfg *config.Config) gin.HandlerFunc { return SaveCompany(cfg) }

func UserProjectNew(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	if user.VerifyStatus != "approved" {
		data := userCenterBase(c, user)
		data["ActiveNav"] = "project"
		data["ActiveSub"] = "new"
		data["error"] = "请先完成企业认证后再发布项目"
		renderUserPage(c, "添加项目", "project-create-content", data)
		return
	}
	data := userCenterBase(c, user)
	data["ActiveNav"] = "project"
	data["ActiveSub"] = "new"
	renderUserPage(c, "添加项目", "project-create-content", data)
}

func UserProductNew(c *gin.Context) { UserProjectNew(c) }

func CreateProject(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	if user.VerifyStatus != "approved" {
		c.Redirect(302, "/my/projects/new?error=verify")
		return
	}

	categoryID, _ := strconv.ParseUint(c.PostForm("category_id"), 10, 64)
	if !categoryAllowedForUser(user, uint(categoryID)) {
		renderProjectFormError(c, user, "所选分类与您的身份不匹配")
		return
	}

	regionsJSON, err := data.BuildRegionsJSON(
		user.Identity,
		c.PostFormArray("regions"),
		c.PostForm("province"),
		c.PostForm("city"),
	)
	if err != nil {
		renderProjectFormError(c, user, err.Error())
		return
	}

	name := strings.TrimSpace(c.PostForm("name"))
	if name == "" {
		renderProjectFormError(c, user, "请填写项目名称")
		return
	}

	project := model.Project{
		UserID:     user.ID,
		CategoryID: uint(categoryID),
		Name:       name,
		ImagePath:  strings.TrimSpace(c.PostForm("image_path")),
		Regions:    regionsJSON,
		Intro:      strings.TrimSpace(c.PostForm("intro")),
		PeriodUnit: "month",
		Status:     "pending",
	}

	if user.Identity == identity.Funder {
		amountWan, _ := strconv.ParseFloat(c.PostForm("amount_wan"), 64)
		ratePercent, _ := strconv.ParseFloat(c.PostForm("rate_percent"), 64)
		periodCount, _ := strconv.Atoi(c.PostForm("period_count"))
		rateType := c.PostForm("rate_type")
		repayMethod := c.PostForm("repay_method")
		if amountWan <= 0 || ratePercent <= 0 || periodCount <= 0 {
			renderProjectFormError(c, user, "请填写完整的金融信息（额度、利率、期数）")
			return
		}
		if rateType != "daily" && rateType != "yearly" {
			rateType = "yearly"
		}
		if repayMethod != "equal_installment" && repayMethod != "equal_principal" {
			repayMethod = "equal_installment"
		}
		project.AmountWan = &amountWan
		project.RatePercent = &ratePercent
		project.PeriodCount = &periodCount
		project.RateType = &rateType
		project.RepayMethod = &repayMethod
	}

	database.DB().Create(&project)
	c.Redirect(302, "/my/projects?created=1")
}

func CreateProduct(c *gin.Context) { CreateProject(c) }

func renderProjectFormError(c *gin.Context, user *model.User, msg string) {
	data := userCenterBase(c, user)
	data["ActiveNav"] = "project"
	data["ActiveSub"] = "new"
	data["error"] = msg
	renderUserPage(c, "添加项目", "project-create-content", data)
}

func UserProjectList(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 10
	status := c.Query("status")

	query := database.DB().Model(&model.Project{}).Where("user_id = ?", user.ID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	var total int64
	query.Count(&total)

	var projects []model.Project
	query.Preload("Category.Parent").Preload("User").Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).Find(&projects)

	data := userCenterBase(c, user)
	data["ActiveNav"] = "project"
	data["ActiveSub"] = "list"
	data["projects"] = projects
	data["products"] = projects
	data["page"] = page
	data["totalPages"] = int(math.Ceil(float64(total) / float64(pageSize)))
	data["status"] = status
	data["created"] = c.Query("created") == "1"
	renderUserPage(c, "项目列表", "project-list-content", data)
}

func UserProductList(c *gin.Context) { UserProjectList(c) }

func DeleteProject(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}
	db := database.DB()
	var project model.Project
	if err := db.Where("user_id = ? AND id = ?", user.ID, c.Param("id")).First(&project).Error; err != nil {
		c.JSON(404, gin.H{"error": "项目不存在"})
		return
	}
	if project.Status != "pending" && project.Status != "rejected" {
		c.JSON(400, gin.H{"error": "仅待审核或已驳回的项目可删除"})
		return
	}
	db.Delete(&project)
	c.Redirect(302, "/my/projects")
}

func DeleteProduct(c *gin.Context) { DeleteProject(c) }

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

func IsSelectedRegion(regions []string, name string) bool { return containsString(regions, name) }
func IsSelectedCategory(ids []uint, id uint) bool         { return containsUint(ids, id) }

func FormatRegionsForUser(raw, userIdentity string) string {
	parts := data.FormatRegionsDisplay(raw, userIdentity)
	return strings.Join(parts, "、")
}
