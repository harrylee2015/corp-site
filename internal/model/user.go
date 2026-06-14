package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Phone          string    `gorm:"type:varchar(11);uniqueIndex;not null" json:"phone"`
	PasswordHash   string    `gorm:"type:varchar(255);not null" json:"-"`
	Role           string    `gorm:"type:varchar(10);not null;default:user" json:"role"`
	RealName       string    `gorm:"type:varchar(50)" json:"real_name"`
	Identity       string    `gorm:"type:varchar(20);not null;default:demander" json:"identity"`
	Nickname       string    `gorm:"type:varchar(50)" json:"nickname"`
	Company        string    `gorm:"type:varchar(100)" json:"company"`
	VerifyStatus   string    `gorm:"type:varchar(15);not null;default:none" json:"verify_status"` // none/approved（上传照片即 approved）
	VerifyDocPath  string    `gorm:"type:varchar(500)" json:"verify_doc_path"`
	Status         string    `gorm:"type:varchar(10);not null;default:active" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (u *User) DisplayName() string {
	if u.RealName != "" {
		return u.RealName
	}
	if u.Nickname != "" {
		return u.Nickname
	}
	if len(u.Phone) >= 7 {
		return u.Phone[:3] + "****" + u.Phone[7:]
	}
	return u.Phone
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}
