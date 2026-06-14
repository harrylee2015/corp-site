package handler

import (
	"bytes"
	"html/template"

	"corp-site/internal/database"
	"corp-site/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var tmpl *template.Template

func SetTemplate(t *template.Template) {
	tmpl = t
}

func renderPage(c *gin.Context, layout, title, contentName string, data gin.H) {
	if userIDStr, exists := c.Get("user_id"); exists {
		data["IsLoggedIn"] = true
		data["UserRole"] = c.GetString("role")
		if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
			var user model.User
			if database.DB().First(&user, "id = ?", userID).Error == nil {
				displayName := user.DisplayName()
				data["UserDisplayName"] = displayName
			}
		}
	}

	categoryID := c.Query("category_id")
	activeCatID, activeParentID, activeCatName := resolveActiveCategory(categoryID)
	data["NavCategories"] = LoadCategoryNav()
	data["CategoryID"] = categoryID
	data["ActiveCategoryID"] = activeCatID
	data["ActiveParentID"] = activeParentID
	data["ActiveCategoryName"] = activeCatName
	data["Keyword"] = c.Query("keyword")
	if tmpl == nil {
		c.String(500, "template not initialized")
		return
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, contentName, data); err != nil {
		c.String(500, "render error: "+err.Error())
		return
	}
	data["Title"] = title
	data["BodyContent"] = template.HTML(buf.String())
	c.HTML(200, layout, data)
}

func renderUserPage(c *gin.Context, title, contentName string, data gin.H) {
	if userIDStr, exists := c.Get("user_id"); exists {
		data["IsLoggedIn"] = true
		data["UserRole"] = c.GetString("role")
		if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
			var user model.User
			if database.DB().First(&user, "id = ?", userID).Error == nil {
				data["UserDisplayName"] = user.DisplayName()
			}
		}
	}
	if tmpl == nil {
		c.String(500, "template not initialized")
		return
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, contentName, data); err != nil {
		c.String(500, "render error: "+err.Error())
		return
	}
	data["Title"] = title
	data["BodyContent"] = template.HTML(buf.String())
	c.HTML(200, "layout/user.html", data)
}
