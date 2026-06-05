package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SmsLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Phone     string    `gorm:"type:varchar(11);not null;index:idx_phone_created" json:"phone"`
	Code      string    `gorm:"type:varchar(6);not null" json:"code"`
	Scene     string    `gorm:"type:varchar(20);not null" json:"scene"`
	ExpiredAt time.Time `gorm:"not null" json:"expired_at"`
	Used      bool      `gorm:"not null;default:false" json:"used"`
	CreatedAt time.Time `gorm:"not null;index:idx_phone_created" json:"created_at"`
}

func (s *SmsLog) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
