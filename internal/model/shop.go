package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Shop struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	ShopName    string    `gorm:"type:varchar(100)" json:"shop_name"`
	Regions     string    `gorm:"type:text" json:"regions"`      // JSON array of province names
	CategoryIDs string    `gorm:"type:text" json:"category_ids"` // JSON array of leaf category ids
	Contact     string    `gorm:"type:varchar(50)" json:"contact"`
	Phone       string    `gorm:"type:varchar(11)" json:"phone"`
	Tel         string    `gorm:"type:varchar(20)" json:"tel"`
	Address     string    `gorm:"type:varchar(200)" json:"address"`
	Intro       string    `gorm:"type:text" json:"intro"`
	BannerPath  string    `gorm:"type:varchar(500)" json:"banner_path"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (s *Shop) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
