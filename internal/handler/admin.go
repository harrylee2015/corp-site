package handler

import (
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

func AdminDashboard(c *gin.Context) {
	db := database.DB()
	var userCount, postCount, pendingCount, todayCount int64
	db.Model(&model.User{}).Where("role = ?", "user").Count(&userCount)
	db.Model(&model.Post{}).Count(&postCount)
	db.Model(&model.Post{}).Where("status = ?", "pending").Count(&pendingCount)
	db.Model(&model.Post{}).Where("created_at >= ?", time.Now().Truncate(24*time.Hour)).Count(&todayCount)

	renderPage(c, "layout/admin.html", "管理后台", "admindash-content", gin.H{
		"userCount":    userCount,
		"postCount":    postCount,
		"pendingCount": pendingCount,
		"todayCount":   todayCount,
		"csrf_token":   c.GetString("csrf_token"),
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

	var categories []model.Category
	db.Order("sort_order ASC").Find(&categories)

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	renderPage(c, "layout/admin.html", "信息管理", "adminposts-content", gin.H{
		"posts":      posts,
		"categories": categories,
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

func AdminExportPage(c *gin.Context) {
	db := database.DB()
	var categories []model.Category
	db.Order("sort_order ASC").Find(&categories)

	renderPage(c, "layout/admin.html", "导出报表", "adminexport-content", gin.H{
		"categories": categories,
		"csrf_token": c.GetString("csrf_token"),
	})
}

func AdminUsers(c *gin.Context) {
	db := database.DB()
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 20
	keyword := c.Query("keyword")

	query := db.Model(&model.User{}).Where("role = ?", "user")
	if keyword != "" {
		query = query.Where("phone ILIKE ? OR nickname ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
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

func MaskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:3] + strings.Repeat("*", 4) + phone[7:]
}
