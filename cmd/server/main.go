package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"corp-site/internal/config"
	"corp-site/internal/database"
	"corp-site/internal/handler"
	"corp-site/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	cfgPath := "config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database.Init(&cfg.Database)
	if err := database.AutoMigrate(); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	database.Seed()

	absUploadPath, _ := filepath.Abs(cfg.Upload.Path)
	if err := os.MkdirAll(absUploadPath, 0755); err != nil {
		log.Fatalf("create upload dir: %v", err)
	}
	cfg.Upload.Path = absUploadPath

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	r.SetFuncMap(template.FuncMap{
		"iterate": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
		"formatFileSize": func(size int64) string {
			if size < 1024 {
				return fmt.Sprintf("%d B", size)
			} else if size < 1048576 {
				return fmt.Sprintf("%.1f KB", float64(size)/1024)
			}
			return fmt.Sprintf("%.1f MB", float64(size)/1048576)
		},
		"MaskPhone": handler.MaskPhone,
		"MaskName":  handler.MaskName,
	})

	t, err := loadTemplates("web/templates", r.FuncMap)
	if err != nil {
		log.Fatalf("load templates: %v", err)
	}
	r.SetHTMLTemplate(t)
	handler.SetTemplate(t)

	secure := cfg.Server.Mode == "release"
	jwtMW := middleware.NewJWTMiddleware(cfg.JWT.Secret, secure)

	r.Static("/static", "./web/static")
	r.GET("/uploads/*filepath", jwtMW.OptionalAuth(), handler.ServeUpload(cfg))

	loginLimiter := middleware.NewRateLimiter(time.Minute, 10)

	// == health ==
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	// == public (CSRF applied) ==
	pub := r.Group("", jwtMW.OptionalAuth(), middleware.CSRFToken())
	{
		pub.GET("/", handler.Index)
		pub.GET("/posts/:id", handler.PostDetail)
		pub.GET("/login", handler.UserLoginPage)
		pub.GET("/register", handler.UserRegisterPage)
		pub.POST("/api/sms/send", handler.SendSMS(cfg))
		pub.POST("/api/auth/register", handler.Register(cfg))
		pub.POST("/api/auth/login", handler.Login(cfg, loginLimiter))
		pub.GET("/logout", handler.Logout(jwtMW))
		pub.GET("/api/posts/list", handler.PostList)
	}

	// == user (JWT required + CSRF) ==
	userGroup := r.Group("", jwtMW.AuthRequired("user"), middleware.CSRFToken())
	{
		userGroup.GET("/my", handler.UserCenterHome)
		userGroup.GET("/my/shop", handler.UserShopPage)
		userGroup.POST("/api/my/shop", handler.SaveShop(cfg))
		userGroup.GET("/my/products/new", handler.UserProductNew)
		userGroup.POST("/api/my/products", handler.CreateProduct)
		userGroup.GET("/my/products", handler.UserProductList)
		userGroup.POST("/api/my/products/:id/delete", handler.DeleteProduct)
		userGroup.GET("/my/profile", handler.UserProfilePage)
		userGroup.POST("/api/my/password", handler.ChangeUserPassword(cfg))
		userGroup.POST("/api/my/verify", handler.UploadVerifyDoc(cfg))
		userGroup.GET("/my/posts", handler.MyPosts)
		userGroup.GET("/my/posts/new", handler.NewPost)
		userGroup.POST("/api/posts", handler.CreatePost(cfg))
		userGroup.GET("/api/posts/:id", handler.MyPostDetail)
		userGroup.POST("/api/posts/:id/delete", handler.DeletePost(cfg))
		userGroup.POST("/api/posts/:id/toggle-list", handler.ToggleListStatus)
		userGroup.POST("/api/upload", handler.UploadFile(cfg))
	}

	// == admin pages (standalone) ==
	adminPub := r.Group("", middleware.CSRFToken())
	{
		adminPub.GET("/admin/login", handler.AdminLoginPage)
		adminPub.POST("/api/admin/login", handler.AdminLogin(cfg, loginLimiter))
	}

	// == admin (JWT required + CSRF) ==
	adminGroup := r.Group("", jwtMW.AuthRequired("admin"), middleware.CSRFToken())
	{
		adminGroup.GET("/admin", handler.AdminDashboard)
		adminGroup.GET("/admin/review", handler.AdminReview)
		adminGroup.GET("/admin/posts", handler.AdminPosts)
		adminGroup.GET("/admin/export", handler.AdminExportPage)
		adminGroup.GET("/admin/users", handler.AdminUsers)
		adminGroup.GET("/admin/password", handler.AdminPasswordPage)
		adminGroup.GET("/api/admin/posts/:id", handler.AdminPostDetail)
		adminGroup.POST("/api/admin/products/:id/review", handler.ReviewProduct)
		adminGroup.GET("/admin/product-review", handler.AdminProductReview)
		adminGroup.POST("/api/admin/posts/:id/delete", handler.AdminDeletePost(cfg))
		adminGroup.POST("/api/admin/password", handler.ChangePassword)
		adminGroup.GET("/api/admin/export", handler.ExportExcel)
		adminGroup.GET("/api/admin/export/preview", handler.ExportPreview)
		adminGroup.GET("/api/admin/users/export", handler.ExportUsersExcel)
		adminGroup.PUT("/api/admin/users/:id/status", handler.UpdateUserStatus)
		adminGroup.PUT("/api/admin/users/:id/verify", handler.UpdateUserVerify)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		fmt.Printf("[Server] starting on http://localhost:%d\n", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("\n[Server] shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server shutdown: %v", err)
	}
	fmt.Println("[Server] stopped")
}

func loadTemplates(root string, funcMap template.FuncMap) (*template.Template, error) {
	t := template.New("").Funcs(funcMap)

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		_, err = t.New(rel).Parse(string(content))
		return err
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}
