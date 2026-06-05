package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID            uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	CategoryID    uint       `gorm:"not null;index" json:"category_id"`
	Title         string     `gorm:"type:varchar(200);not null" json:"title"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	Contact       string     `gorm:"type:varchar(100)" json:"contact"`
	ContactPhone  string     `gorm:"type:varchar(11)" json:"contact_phone"`
	Status        string     `gorm:"type:varchar(15);not null;default:pending;index" json:"status"`
	RejectReason  string     `gorm:"type:text" json:"reject_reason,omitempty"`
	ReviewedBy    *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	User        User         `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category    Category     `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Attachments []Attachment `gorm:"foreignKey:PostID" json:"attachments,omitempty"`
	Reviewer    *User        `gorm:"foreignKey:ReviewedBy" json:"reviewer,omitempty"`
}

func (p *Post) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}
