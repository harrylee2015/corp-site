package handler

import (
	"github.com/gin-gonic/gin"
)

func UserLoginPage(c *gin.Context) {
	renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{
		"csrf_token": c.GetString("csrf_token"),
	})
}

func UserRegisterPage(c *gin.Context) {
	renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{
		"csrf_token": c.GetString("csrf_token"),
	})
}

func AdminLoginPage(c *gin.Context) {
	renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{
		"csrf_token": c.GetString("csrf_token"),
	})
}
