package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID   uint       `gorm:"not null;index" json:"category_id"`
	Name         string     `gorm:"type:varchar(200);not null" json:"name"`
	ImagePath    string     `gorm:"type:varchar(500)" json:"image_path"`
	AmountWan    float64    `gorm:"not null" json:"amount_wan"`
	RateType     string     `gorm:"type:varchar(10);not null" json:"rate_type"` // daily | yearly
	RatePercent  float64    `gorm:"not null" json:"rate_percent"`
	PeriodCount  int        `gorm:"not null" json:"period_count"`
	PeriodUnit   string     `gorm:"type:varchar(10);not null;default:month" json:"period_unit"` // month
	RepayMethod  string     `gorm:"type:varchar(30);not null" json:"repay_method"`              // equal_installment | equal_principal
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

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
