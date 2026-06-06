package handler

import (
	"bytes"
	"html/template"

	"github.com/gin-gonic/gin"
)

var tmpl *template.Template

func SetTemplate(t *template.Template) {
	tmpl = t
}

func renderPage(c *gin.Context, layout, title, contentName string, data gin.H) {
	if _, exists := c.Get("user_id"); exists {
		data["IsLoggedIn"] = true
		data["UserRole"] = c.GetString("role")
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
	c.HTML(200, layout, data)
}
