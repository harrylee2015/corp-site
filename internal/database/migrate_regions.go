package database

import (
	"encoding/json"
	"log"

	"corp-site/internal/data"
	"corp-site/internal/identity"
	"corp-site/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MigrateRegionFormats 非资金方：旧多选省 JSON 数组格式不再适用，清空以便用户重选省市
func MigrateRegionFormats() {
	db := DB()
	if db == nil {
		return
	}

	var users []model.User
	db.Where("role = ? AND identity IN ?", "user", []string{identity.Demander, identity.Supplier}).Find(&users)
	for _, u := range users {
		fixCompanyRegions(db, u.ID)
		fixProjectRegions(db, u.ID)
	}
}

func fixCompanyRegions(db *gorm.DB, userID uuid.UUID) {
	var company model.Company
	if db.Where("user_id = ?", userID).First(&company).Error != nil {
		return
	}
	if company.Regions == "" {
		return
	}
	if isProvinceArray(company.Regions) {
		db.Model(&company).Update("regions", "")
		log.Printf("[DB] cleared legacy region array for company user %s", userID)
	}
}

func fixProjectRegions(db *gorm.DB, userID uuid.UUID) {
	var projects []model.Project
	db.Where("user_id = ?", userID).Find(&projects)
	for _, p := range projects {
		if p.Regions == "" {
			continue
		}
		if isProvinceArray(p.Regions) {
			db.Model(&p).Update("regions", "")
		}
	}
}

func isProvinceArray(raw string) bool {
	var list []string
	if json.Unmarshal([]byte(raw), &list) != nil || len(list) == 0 {
		return false
	}
	var obj data.SingleCityRegion
	if json.Unmarshal([]byte(raw), &obj) == nil && obj.Province != "" {
		return false
	}
	return true
}
