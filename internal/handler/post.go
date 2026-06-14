package handler

import (
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"corp-site/internal/config"
	"corp-site/internal/database"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func PostDetail(c *gin.Context) {
	db := database.DB()
	var post model.Post
	if err := db.Preload("Category.Parent").Preload("User").
		Preload("Attachments").First(&post, "id = ?", c.Param("id")).Error; err != nil {
		renderPage(c, "layout/base.html", "信息不存在", "postdetail-content", gin.H{"error": "信息不存在", "csrf_token": c.GetString("csrf_token")})
		return
	}

	visible := post.Status == "approved"
	if !visible {
		userIDStr, _ := c.Get("user_id")
		role, _ := c.Get("role")
		if userIDStr != nil && (userIDStr.(string) == post.UserID.String() || role == "admin") {
			visible = true
		}
	}
	if !visible {
		renderPage(c, "layout/base.html", "信息不存在", "postdetail-content", gin.H{"error": "信息已下架", "csrf_token": c.GetString("csrf_token")})
		return
	}

	userIDStr, _ := c.Get("user_id")
	role, _ := c.Get("role")
	isOwner := userIDStr != nil && userIDStr.(string) == post.UserID.String()
	isAdmin := role == "admin"
	showPrivate := isOwner || isAdmin

	renderPage(c, "layout/base.html", "信息详情", "postdetail-content", gin.H{
		"post":          post,
		"showPrivate":   showPrivate,
		"maskedContact": MaskName(post.Contact),
		"maskedPhone":   MaskPhone(post.ContactPhone),
		"csrf_token":    c.GetString("csrf_token"),
	})
}

func MyPosts(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.Redirect(302, "/login")
		return
	}
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20

	var total int64
	db.Model(&model.Post{}).Where("user_id = ?", user.ID).Count(&total)

	var posts []model.Post
	db.Where("user_id = ?", user.ID).Preload("Category").Preload("Attachments").
		Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	flash, _ := c.Cookie("flash")
	if flash != "" {
		c.SetCookie("flash", "", -1, "/", "", false, false)
	}

	renderPage(c, "layout/base.html", "我的发布", "dashboard-content", gin.H{
		"posts":      posts,
		"page":       page,
		"totalPages": totalPages,
		"flash":      flash,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func NewPost(c *gin.Context) {
	renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
		"navCategories": LoadCategoryNav(),
		"csrf_token":    c.GetString("csrf_token"),
	})
}

func CreatePost(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.JSON(401, gin.H{"error": "请先登录"})
			return
		}

		var req struct {
			CategoryID   uint   `form:"category_id" binding:"required"`
			Title        string `form:"title" binding:"required,min=5,max=100"`
			Content      string `form:"content" binding:"required"`
			Contact      string `form:"contact"`
			ContactPhone string `form:"contact_phone"`
			AttachIDs    string `form:"attach_ids"`
		}
		if err := c.ShouldBind(&req); err != nil {
			errMsg := "请填写完整信息（分类、标题5-100字、内容必填）"
			candidate := err.Error()
			if strings.Contains(candidate, "category_id") {
				errMsg = "请选择信息分类"
			} else if strings.Contains(candidate, "Title") {
				if strings.Contains(candidate, "min") {
					errMsg = "标题至少需要5个字"
				} else {
					errMsg = "标题需5-100字"
				}
			} else if strings.Contains(candidate, "Content") {
				errMsg = "请输入信息内容"
			}
			renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
				"error":         errMsg,
				"navCategories": LoadCategoryNav(),
				"csrf_token":    c.GetString("csrf_token"),
			})
			return
		}

		if req.ContactPhone != "" && len(req.ContactPhone) != 11 {
			renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
				"error":         "联系电话格式不正确，需为11位数字",
				"navCategories": LoadCategoryNav(),
				"csrf_token":    c.GetString("csrf_token"),
			})
			return
		}

		db := database.DB()
		post := model.Post{
			UserID:       user.ID,
			CategoryID:   req.CategoryID,
			Title:        req.Title,
			Content:      req.Content,
			Contact:      req.Contact,
			ContactPhone: req.ContactPhone,
			Status:       "pending",
		}

		tx := db.Begin()
		if tx.Error != nil {
			renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
				"error":         "发布失败，请重试",
				"navCategories": LoadCategoryNav(),
				"csrf_token":    c.GetString("csrf_token"),
			})
			return
		}

		if err := tx.Create(&post).Error; err != nil {
			tx.Rollback()
			renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
				"error":         "发布失败，请重试",
				"navCategories": LoadCategoryNav(),
				"csrf_token":    c.GetString("csrf_token"),
			})
			return
		}

		if req.AttachIDs != "" {
			ids := strings.Split(req.AttachIDs, ",")
			for _, idStr := range ids {
				attachID, err := uuid.Parse(strings.TrimSpace(idStr))
				if err != nil {
					continue
				}
				if err := tx.Model(&model.Attachment{}).Where("id = ? AND post_id IS NULL", attachID).
					Update("post_id", post.ID).Error; err != nil {
					tx.Rollback()
					renderPage(c, "layout/base.html", "发布信息", "postcreate-content", gin.H{
						"error":         "发布失败，请重试",
						"navCategories": LoadCategoryNav(),
						"csrf_token":    c.GetString("csrf_token"),
					})
					return
				}
			}
		}

		tx.Commit()
		c.SetCookie("flash", "信息已提交，请等待审核", 5, "/", "", false, false)
		c.Redirect(302, "/my/posts")
	}
}

func MyPostDetail(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}
	db := database.DB()
	var post model.Post
	if err := db.Where("user_id = ?", user.ID).Preload("Category").Preload("Attachments").
		First(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "信息不存在"})
		return
	}
	c.JSON(200, gin.H{"post": post})
}

func DeletePost(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := authUserFromContext(c)
		if user == nil {
			c.JSON(401, gin.H{"error": "请先登录"})
			return
		}
		db := database.DB()
		var post model.Post
		if err := db.Where("user_id = ? AND status = ?", user.ID, "pending").
			First(&post, "id = ?", c.Param("id")).Error; err != nil {
			c.JSON(400, gin.H{"error": "只能删除待审核的信息"})
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
		c.Redirect(302, "/my/posts")
	}
}

func UploadFile(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "请选择文件"})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		ext = strings.TrimPrefix(ext, ".")
		allowed := strings.Split(cfg.Upload.AllowedTypes, ",")
		allowedMap := make(map[string]bool)
		for _, t := range allowed {
			allowedMap[strings.TrimSpace(t)] = true
		}
		if !allowedMap[ext] {
			c.JSON(400, gin.H{"error": "不支持的文件类型"})
			return
		}

		if file.Size > cfg.Upload.MaxSize {
			c.JSON(400, gin.H{"error": "文件大小超过限制"})
			return
		}

		monthDir := time.Now().Format("200601")
		uploadDir := filepath.Join(cfg.Upload.Path, monthDir)
		os.MkdirAll(uploadDir, 0755)

		newName := fmt.Sprintf("%s.%s", uuid.New().String(), ext)
		savePath := filepath.Join(uploadDir, newName)

		src, err := file.Open()
		if err != nil {
			c.JSON(500, gin.H{"error": "文件读取失败"})
			return
		}
		defer src.Close()

		dst, err := os.Create(savePath)
		if err != nil {
			c.JSON(500, gin.H{"error": "文件保存失败"})
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			c.JSON(500, gin.H{"error": "文件写入失败"})
			return
		}

		relPath := filepath.Join(monthDir, newName)
		db := database.DB()
		attach := model.Attachment{
			FileName: file.Filename,
			FilePath: relPath,
			FileSize: file.Size,
		}
		db.Create(&attach)

		publicURL := fmt.Sprintf("/uploads/%s", relPath)
		c.JSON(200, gin.H{
			"id":       attach.ID,
			"filename": file.Filename,
			"url":      publicURL,
			"size":     file.Size,
		})
	}
}

func PostList(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 12
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	query := db.Model(&model.Post{}).Where("status = ?", "approved").
		Preload("Category.Parent").Preload("User").Preload("Attachments")

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if keyword != "" {
		for _, word := range strings.Fields(keyword) {
			like := "%" + word + "%"
			query = query.Where("(title ILIKE ? OR content ILIKE ? OR contact ILIKE ?)", like, like, like)
		}
	}

	var total int64
	query.Count(&total)

	var posts []model.Post
	query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&posts)

	hasMore := int64(page*pageSize) < total

	type row struct {
		ID           string `json:"id"`
		Title        string `json:"title"`
		Content      string `json:"content"`
		CategoryName string `json:"category"`
		CategoryCls  string `json:"category_cls"`
		Nickname     string `json:"nickname"`
		Date         string `json:"date"`
		HasAttach    bool   `json:"has_attach"`
		AttachCount  int    `json:"attach_count"`
	}

	rows := make([]row, len(posts))
	for i, p := range posts {
		nickname := MaskName(p.User.DisplayName())
		if nickname == "" {
			nickname = MaskPhone(p.User.Phone)
		}
		cls := catColorClass(catParentName(p.Category))
		rows[i] = row{
			ID:           p.ID.String(),
			Title:        p.Title,
			Content:      p.Content,
			CategoryName: formatPostCategory(p.Category),
			CategoryCls:  cls,
			Nickname:     nickname,
			Date:         p.CreatedAt.Format("01-02"),
			HasAttach:    false,
			AttachCount:  0,
		}
	}

	c.JSON(200, gin.H{"posts": rows, "has_more": hasMore})
}

func ToggleListStatus(c *gin.Context) {
	user := authUserFromContext(c)
	if user == nil {
		c.JSON(401, gin.H{"error": "请先登录"})
		return
	}
	db := database.DB()
	var post model.Post
	if err := db.Where("user_id = ?", user.ID).First(&post, "id = ?", c.Param("id")).Error; err != nil {
		c.JSON(404, gin.H{"error": "信息不存在"})
		return
	}
	if post.Status == "approved" {
		db.Model(&post).Update("status", "delisted")
		c.JSON(200, gin.H{"message": "已下架", "status": "delisted"})
	} else if post.Status == "delisted" {
		db.Model(&post).Update("status", "approved")
		c.JSON(200, gin.H{"message": "已上架", "status": "approved"})
	} else {
		c.JSON(400, gin.H{"error": "当前状态不支持此操作"})
	}
}
