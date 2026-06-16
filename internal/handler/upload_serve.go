package handler

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"corp-site/internal/config"
	"corp-site/internal/database"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func ServeUpload(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		relPath := strings.TrimPrefix(c.Param("filepath"), "/")
		relPath = filepath.ToSlash(relPath)
		if relPath == "" || strings.Contains(relPath, "..") {
			c.Status(http.StatusNotFound)
			return
		}

		absPath := filepath.Clean(filepath.Join(cfg.Upload.Path, filepath.FromSlash(relPath)))
		uploadRoot := filepath.Clean(cfg.Upload.Path)
		if !strings.HasPrefix(absPath, uploadRoot+string(os.PathSeparator)) && absPath != uploadRoot {
			c.Status(http.StatusForbidden)
			return
		}
		if _, err := os.Stat(absPath); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		role, _ := c.Get("role")
		if role == "admin" {
			c.File(absPath)
			return
		}

		userIDStr, hasUser := c.Get("user_id")
		if !hasUser {
			c.Status(http.StatusForbidden)
			return
		}
		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.Status(http.StatusForbidden)
			return
		}

		if canAccessUpload(userID, relPath) {
			c.File(absPath)
			return
		}
		c.Status(http.StatusForbidden)
	}
}

func normalizeUploadRelPath(path string) string {
	path = strings.TrimSpace(path)
	path = filepath.ToSlash(path)
	if path == "" || strings.Contains(path, "..") {
		return ""
	}
	return path
}

func userOwnsUpload(userID uuid.UUID, relPath string) bool {
	norm := normalizeUploadRelPath(relPath)
	if norm == "" {
		return false
	}
	var attach model.Attachment
	return database.DB().Where("file_path = ? AND user_id = ?", norm, userID).First(&attach).Error == nil
}

func canAccessUpload(userID uuid.UUID, relPath string) bool {
	db := database.DB()
	norm := normalizeUploadRelPath(relPath)
	if norm == "" {
		return false
	}

	var attach model.Attachment
	if err := db.Where("file_path = ?", norm).First(&attach).Error; err == nil {
		if attach.UserID != uuid.Nil && attach.UserID == userID {
			return true
		}
		if attach.PostID != nil {
			var post model.Post
			if db.First(&post, "id = ?", *attach.PostID).Error == nil {
				return post.UserID == userID
			}
		}
		return false
	}

	var user model.User
	if db.First(&user, "id = ?", userID).Error == nil && user.VerifyDocPath == norm {
		return true
	}

	var company model.Company
	if db.Where("user_id = ? AND banner_path = ?", userID, norm).First(&company).Error == nil {
		return true
	}

	var project model.Project
	if db.Where("user_id = ? AND image_path = ?", userID, norm).First(&project).Error == nil {
		return true
	}

	return false
}
