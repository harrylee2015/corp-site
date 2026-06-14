package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Company 公司信息
type Company struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	ShopName       string    `gorm:"type:varchar(100)" json:"shop_name"` // UI 显示为公司名称
	EstablishedAt  string    `gorm:"type:varchar(7)" json:"established_at"` // YYYY-MM
	Regions        string    `gorm:"type:text" json:"regions"`
	CategoryIDs    string    `gorm:"type:text" json:"category_ids"`
	Contact        string    `gorm:"type:varchar(50)" json:"contact"`
	Phone          string    `gorm:"type:varchar(11)" json:"phone"`
	Tel            string    `gorm:"type:varchar(20)" json:"tel"`
	Address        string    `gorm:"type:varchar(200)" json:"address"`
	Intro          string    `gorm:"type:text" json:"intro"`
	BannerPath     string    `gorm:"type:varchar(500)" json:"banner_path"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Company) TableName() string { return "companies" }

func (c *Company) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// Shop 兼容旧代码引用
type Shop = Company
