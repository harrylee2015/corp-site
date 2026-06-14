package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Project 项目
type Project struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID   uint       `gorm:"not null;index" json:"category_id"`
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	ImagePath    string     `gorm:"type:varchar(500)" json:"image_path"`
	AmountWan    *float64   `json:"amount_wan,omitempty"`
	RateType     *string    `gorm:"type:varchar(10)" json:"rate_type,omitempty"`
	RatePercent  *float64   `json:"rate_percent,omitempty"`
	PeriodCount  *int       `json:"period_count,omitempty"`
	PeriodUnit   string     `gorm:"type:varchar(10);default:month" json:"period_unit"`
	RepayMethod  *string    `gorm:"type:varchar(30)" json:"repay_method,omitempty"`
	Regions      string     `gorm:"type:text" json:"regions"`
	Intro        string     `gorm:"type:text" json:"intro"`
	Status       string     `gorm:"type:varchar(15);not null;default:pending;index" json:"status"`
	RejectReason string     `gorm:"type:text" json:"reject_reason,omitempty"`
	ReviewedBy   *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt   *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Project) TableName() string { return "projects" }

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

func (p *Project) IsFunderProject() bool {
	return p.AmountWan != nil && *p.AmountWan > 0
}

// Product 兼容旧代码引用
type Product = Project
