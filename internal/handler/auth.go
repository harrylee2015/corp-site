package handler

import (
	"time"

	"corp-site/internal/config"
	"corp-site/internal/database"
	"corp-site/internal/middleware"
	"corp-site/internal/model"
	"corp-site/internal/sms"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Login(cfg *config.Config, limiter *middleware.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "登录尝试过于频繁，请稍后再试", "csrf_token": c.GetString("csrf_token")})
			return
		}

		var req struct {
			Phone    string `json:"phone" form:"phone" binding:"required"`
			Password string `json:"password" form:"password" binding:"required"`
		}
		token := c.GetString("csrf_token")
		if err := c.ShouldBind(&req); err != nil {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "请填写手机号和密码", "csrf_token": token})
			return
		}

		db := database.DB()
		var user model.User
		if err := db.Where("phone = ? AND role = ?", req.Phone, "user").First(&user).Error; err != nil {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "手机号或密码错误", "csrf_token": token})
			return
		}

		if user.Status != "active" {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "账号已被禁用，请联系管理员", "csrf_token": token})
			return
		}

		if !user.CheckPassword(req.Password) {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "手机号或密码错误", "csrf_token": token})
			return
		}

		jwtMW := middleware.NewJWTMiddleware(cfg.JWT.Secret, cfg.Server.Mode == "release")
		accessToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.AccessTTL)
		if err != nil {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "系统错误，请重试", "csrf_token": token})
			return
		}
		refreshToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.RefreshTTL)
		if err != nil {
			renderPage(c, "layout/base.html", "用户登录", "login-content", gin.H{"error": "系统错误，请重试", "csrf_token": token})
			return
		}

		jwtMW.SetTokenCookies(c, accessToken, refreshToken, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
		c.Redirect(302, "/my/posts")
	}
}

func Register(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Phone           string `json:"phone" form:"phone" binding:"required,len=11"`
			Code            string `json:"code" form:"code" binding:"required,len=6"`
			Password        string `json:"password" form:"password" binding:"required,min=8,max=20"`
			PasswordConfirm string `json:"password_confirm" form:"password_confirm" binding:"required"`
			Nickname        string `json:"nickname" form:"nickname"`
			Company         string `json:"company" form:"company"`
		}
		if err := c.ShouldBind(&req); err != nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "请填写完整信息", "csrf_token": c.GetString("csrf_token")})
			return
		}

		if req.Password != req.PasswordConfirm {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "两次密码输入不一致", "csrf_token": c.GetString("csrf_token")})
			return
		}

		hasLetter := false
		hasDigit := false
		for _, ch := range req.Password {
			if ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' {
				hasLetter = true
			}
			if ch >= '0' && ch <= '9' {
				hasDigit = true
			}
		}
		if !hasLetter || !hasDigit {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "密码需包含字母和数字", "csrf_token": c.GetString("csrf_token")})
			return
		}

		if req.Nickname != "" && (len(req.Nickname) < 2 || len(req.Nickname) > 20) {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "昵称需2-20个字符", "csrf_token": c.GetString("csrf_token")})
			return
		}

		db := database.DB()

		var smsLog model.SmsLog
		if err := db.Where("phone = ? AND code = ? AND scene = ? AND used = ? AND expired_at > ?",
			req.Phone, req.Code, "register", false, time.Now()).First(&smsLog).Error; err != nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "验证码错误或已过期", "csrf_token": c.GetString("csrf_token")})
			return
		}

		var exists model.User
		if err := db.Where("phone = ?", req.Phone).First(&exists).Error; err == nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "该手机号已注册", "csrf_token": c.GetString("csrf_token")})
			return
		}

		user := model.User{
			Phone:    req.Phone,
			Role:     "user",
			Nickname: req.Nickname,
			Company:  req.Company,
			Status:   "active",
		}
		if err := user.SetPassword(req.Password); err != nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "系统错误，请重试", "csrf_token": c.GetString("csrf_token")})
			return
		}

		db.Create(&user)
		db.Model(&smsLog).Update("used", true)

		jwtMW := middleware.NewJWTMiddleware(cfg.JWT.Secret, cfg.Server.Mode == "release")
		accessToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.AccessTTL)
		if err != nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "系统错误，请重试", "csrf_token": c.GetString("csrf_token")})
			return
		}
		refreshToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.RefreshTTL)
		if err != nil {
			renderPage(c, "layout/base.html", "用户注册", "register-content", gin.H{"error": "系统错误，请重试", "csrf_token": c.GetString("csrf_token")})
			return
		}

		jwtMW.SetTokenCookies(c, accessToken, refreshToken, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
		c.Redirect(302, "/my/posts")
	}
}

func AdminLogin(cfg *config.Config, limiter *middleware.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow(c.ClientIP()) {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "登录尝试过于频繁，请稍后再试", "csrf_token": c.GetString("csrf_token")})
			return
		}

		var req struct {
			Phone    string `json:"phone" form:"phone" binding:"required"`
			Password string `json:"password" form:"password" binding:"required"`
		}
		token := c.GetString("csrf_token")
		if err := c.ShouldBind(&req); err != nil {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "请填写手机号和密码", "csrf_token": token})
			return
		}

		db := database.DB()
		var user model.User
		if err := db.Where("phone = ? AND role = ?", req.Phone, "admin").First(&user).Error; err != nil {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "手机号或密码错误", "csrf_token": token})
			return
		}

		if user.Status != "active" {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "账号已被禁用", "csrf_token": token})
			return
		}

		if !user.CheckPassword(req.Password) {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "手机号或密码错误", "csrf_token": token})
			return
		}

		jwtMW := middleware.NewJWTMiddleware(cfg.JWT.Secret, cfg.Server.Mode == "release")
		accessToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.AccessTTL)
		if err != nil {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "系统错误，请重试", "csrf_token": token})
			return
		}
		refreshToken, err := jwtMW.GenerateToken(user.ID.String(), user.Role, cfg.JWT.RefreshTTL)
		if err != nil {
			renderPage(c, "layout/admin.html", "管理员登录", "adminlogin-content", gin.H{"error": "系统错误，请重试", "csrf_token": token})
			return
		}

		jwtMW.SetTokenCookies(c, accessToken, refreshToken, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)
		c.Redirect(302, "/admin")
	}
}

func SendSMS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Phone string `json:"phone" form:"phone" binding:"required,len=11"`
			Scene string `json:"scene" form:"scene" binding:"required"`
		}
		if err := c.ShouldBind(&req); err != nil {
			c.JSON(400, gin.H{"error": "参数错误"})
			return
		}

		db := database.DB()

		var minuteCount int64
		db.Model(&model.SmsLog{}).Where("phone = ? AND created_at > ?",
			req.Phone, time.Now().Add(-time.Minute)).Count(&minuteCount)
		if minuteCount >= int64(cfg.SMS.SendLimitPerMinute) {
			c.JSON(429, gin.H{"error": "发送过于频繁，请稍后再试"})
			return
		}

		var hourCount int64
		db.Model(&model.SmsLog{}).Where("phone = ? AND created_at > ?",
			req.Phone, time.Now().Add(-time.Hour)).Count(&hourCount)
		if hourCount >= int64(cfg.SMS.SendLimitPerHour) {
			c.JSON(429, gin.H{"error": "发送次数已达上限，请稍后再试"})
			return
		}

		code := sms.GenerateCode(cfg.SMS.CodeLength)
		expiredAt := time.Now().Add(time.Duration(cfg.SMS.CodeTTL) * time.Second)

		log := model.SmsLog{
			Phone:     req.Phone,
			Code:      code,
			Scene:     req.Scene,
			ExpiredAt: expiredAt,
		}
		db.Create(&log)

		go func() {
			provider := getSMSProvider(cfg)
			_ = provider.Send(req.Phone, code)
		}()

		if cfg.SMS.Mock {
			c.JSON(200, gin.H{"message": "验证码已发送", "code": code})
			return
		}
		c.JSON(200, gin.H{"message": "验证码已发送"})
	}
}

func Logout(jwtMW *middleware.JWTMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtMW.ClearTokenCookies(c)
		c.Redirect(302, "/")
	}
}

func authUserFromContext(c *gin.Context) *model.User {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		return nil
	}
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		return nil
	}
	var user model.User
	database.DB().First(&user, "id = ?", userID)
	return &user
}

func getSMSProvider(cfg *config.Config) sms.Provider {
	if cfg.SMS.Mock {
		return &sms.MockProvider{}
	}
	return sms.NewTencentProvider(
		cfg.SMS.SecretID, cfg.SMS.SecretKey,
		cfg.SMS.SDKAppID, cfg.SMS.SignName, cfg.SMS.TemplateID,
	)
}
