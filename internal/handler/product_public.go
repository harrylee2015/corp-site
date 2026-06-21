package handler

import (
	"fmt"
	"hash/fnv"
	"math"
	"strconv"
	"strings"

	"corp-site/internal/data"
	"corp-site/internal/database"
	"corp-site/internal/identity"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	photos := AssignProjectPhotos(projects, (page-1)*pageSize)

	renderPage(c, "layout/base.html", "首页 - 金筹设备租赁", "index-content", gin.H{
		"projects":   projects,
		"products":   projects,
		"photos":     photos,
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

	photos := AssignProjectPhotos(projects, (page-1)*pageSize)

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
		Photo    string `json:"photo"`
		Hue      int    `json:"hue"`
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
			Photo:    photos[i],
			Hue:      ProjectHue(p.ID),
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
		var childIDs []uint
		db.Model(&model.Category{}).Where("parent_id = ?", categoryID).Pluck("id", &childIDs)
		if len(childIDs) > 0 {
			query = query.Where("category_id IN ?", childIDs)
		} else {
			query = query.Where("category_id = ?", categoryID)
		}
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

const photoBase = "/static/img/photos/"

// 二级分类 -> 图片池（对题、按二级分散）
var subCategoryPhotoPool = map[string][]string{
	"光伏发电":     {"solar-panels.jpg", "solar-field-2.jpg", "solar-sunset.jpg"},
	"储能电站":     {"battery-storage.jpg", "substation.jpg", "power-grid.jpg"},
	"风力发电":     {"wind-farm.jpg", "wind-farm-2.jpg", "wind-farm-3.jpg"},
	"垃圾发电":     {"waste-plant.jpg", "power-plant.jpg", "industrial.jpg"},
	"水利发电":     {"hydro-dam.jpg", "hydro-station.jpg", "power-lines.jpg"},
	"充电桩":      {"ev-charging.jpg", "city-night.jpg", "substation.jpg"},
	"光储充一体化项目": {"solar-panels.jpg", "battery-storage.jpg", "ev-charging.jpg"},
	"央国企设备租赁":  {"office-building.jpg", "business-meeting.jpg", "industrial.jpg"},
	"上市公司设备租赁": {"business-team.jpg", "office-building.jpg", "finance.jpg"},
	"中小微企业设备租赁": {"business.jpg", "business-meeting.jpg", "warehouse.jpg"},
	"金租":       {"finance.jpg", "office-building.jpg", "business.jpg"},
	"商租":       {"business-team.jpg", "business-meeting.jpg", "city-skyline.jpg"},
	"外资":       {"office-building.jpg", "city-skyline.jpg", "business-team.jpg"},
	"地级市平台公司":  {"city-skyline.jpg", "office-building.jpg", "city-night.jpg"},
	"区县级平台公司":  {"office-building.jpg", "city-night.jpg", "business.jpg"},
}

// 一级分类 -> 图片池（无二级时回退）
var categoryPhotoPool = map[string][]string{
	"新能源项目": {"solar-panels.jpg", "wind-farm.jpg", "power-grid.jpg", "solar-field-2.jpg", "battery-storage.jpg", "hydro-station.jpg", "ev-charging.jpg"},
	"企业类项目": {"finance.jpg", "business.jpg", "office-building.jpg", "business-meeting.jpg", "business-team.jpg"},
	"政信类项目": {"city-skyline.jpg", "office-building.jpg", "city-night.jpg", "business.jpg"},
	"租赁公司":  {"industrial.jpg", "heavy-machinery.jpg", "construction.jpg", "forklift.jpg", "warehouse.jpg"},
	"电站出售方": {"power-grid.jpg", "solar-panels.jpg", "substation.jpg", "power-lines.jpg", "industrial.jpg"},
	"电站收购方": {"power-grid.jpg", "solar-panels.jpg", "substation.jpg", "power-lines.jpg", "industrial.jpg"},
	"其他类":   {"business.jpg", "finance.jpg", "city-skyline.jpg", "city-night.jpg", "warehouse.jpg"},
}

var defaultPhotoPool = []string{"business.jpg", "finance.jpg", "solar-panels.jpg", "office-building.jpg"}

// 全局图库（去重借用，保证同页不重复）——不含 hero 大图
var allPhotos = []string{
	"solar-panels.jpg", "solar-field-2.jpg", "solar-sunset.jpg",
	"wind-farm.jpg", "wind-farm-2.jpg", "wind-farm-3.jpg", "battery-storage.jpg",
	"power-grid.jpg", "power-lines.jpg", "substation.jpg", "power-plant.jpg", "waste-plant.jpg",
	"hydro-dam.jpg", "hydro-station.jpg", "ev-charging.jpg",
	"industrial.jpg", "heavy-machinery.jpg", "construction.jpg", "forklift.jpg", "warehouse.jpg",
	"finance.jpg", "business.jpg", "office-building.jpg", "business-meeting.jpg", "business-team.jpg",
	"city-skyline.jpg", "city-night.jpg",
}

func hashID(id uuid.UUID) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id.String()))
	return h.Sum32()
}

// poolForProject 优先用二级分类池，否则回退一级池，再回退默认池。
func poolForProject(p model.Project) []string {
	if p.Category.Parent != nil {
		if pool, ok := subCategoryPhotoPool[p.Category.Name]; ok && len(pool) > 0 {
			return pool
		}
		if pool, ok := categoryPhotoPool[p.Category.Parent.Name]; ok && len(pool) > 0 {
			return pool
		}
	} else if pool, ok := categoryPhotoPool[p.Category.Name]; ok && len(pool) > 0 {
		return pool
	}
	return defaultPhotoPool
}

// ProjectPhoto 单项取图（按二级池 + ID 哈希），用于无需去重的场景。
func ProjectPhoto(p model.Project) string {
	pool := poolForProject(p)
	return photoBase + pool[int(hashID(p.ID)%uint32(len(pool)))]
}

// AssignProjectPhotos 给一页项目分配底图，保证本页内不重复：
// 优先从各自二级池里取“本页未用过”的图；池在本页用尽则从全局图库借未用图。
func AssignProjectPhotos(projects []model.Project, offset int) []string {
	out := make([]string, len(projects))
	used := make(map[string]bool, len(projects))
	for i, p := range projects {
		pool := poolForProject(p)
		chosen := ""
		start := int(hashID(p.ID) % uint32(len(pool)))
		for k := 0; k < len(pool); k++ {
			if cand := pool[(start+k)%len(pool)]; !used[cand] {
				chosen = cand
				break
			}
		}
		if chosen == "" {
			g := (offset + i) % len(allPhotos)
			for k := 0; k < len(allPhotos); k++ {
				if cand := allPhotos[(g+k)%len(allPhotos)]; !used[cand] {
					chosen = cand
					break
				}
			}
		}
		if chosen == "" {
			chosen = pool[start]
		}
		used[chosen] = true
		out[i] = photoBase + chosen
	}
	return out
}

// ProjectHue 给每个项目一个专属色相（0-359），用于封面色块叠加。
func ProjectHue(id uuid.UUID) int {
	return int(hashID(id) % 360)
}
