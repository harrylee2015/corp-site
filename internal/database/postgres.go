package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"corp-site/internal/config"
	"corp-site/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db   *gorm.DB
	once sync.Once
)

func Init(cfg *config.DatabaseConfig) {
	once.Do(func() {
		logLevel := logger.Warn
		if os.Getenv("DB_LOG_LEVEL") == "info" {
			logLevel = logger.Info
		}

		var d *gorm.DB
		var err error

		for i := 0; i < 30; i++ {
			d, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
				Logger: logger.Default.LogMode(logLevel),
			})
			if err == nil {
				break
			}
			if i == 0 {
				fmt.Println("[DB] waiting for PostgreSQL...")
			}
			sleep := time.Duration(i+1) * time.Second
			if sleep > 10*time.Second {
				sleep = 10 * time.Second
			}
			time.Sleep(sleep)
		}
		if err != nil {
			log.Fatalf("connect postgres failed after 30 retries: %v", err)
		}

		sqlDB, err := d.DB()
		if err != nil {
			log.Fatalf("get sql.DB: %v", err)
		}
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)

		db = d
		fmt.Println("[DB] PostgreSQL connected")
	})
}

func DB() *gorm.DB {
	return db
}

func AutoMigrate() error {
	return db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Post{},
		&model.Attachment{},
		&model.SmsLog{},
	)
}

func Seed() {
	SeedCategories()

	var adminCount int64
	db.Model(&model.User{}).Where("role = ?", "admin").Count(&adminCount)
	if adminCount == 0 {
		adminPassword := "Admin@123"
		adminPhone := "13800000000"
		if v := os.Getenv("ADMIN_PASSWORD"); v != "" {
			adminPassword = v
		}
		if v := os.Getenv("ADMIN_PHONE"); v != "" {
			adminPhone = v
		}
		admin := &model.User{
			Phone:        adminPhone,
			PasswordHash: "",
			Role:         "admin",
			Nickname:     "超级管理员",
			Status:       "active",
		}
		admin.SetPassword(adminPassword)
		db.Create(admin)
		fmt.Printf("[DB] default admin seeded (phone: %s)\n", adminPhone)
	}
}
