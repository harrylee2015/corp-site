package database

import (
	"log"
)

// MigrateTableNames 将旧表 shops/products 重命名为 companies/projects（幂等）
func MigrateTableNames() {
	db := DB()
	if db == nil {
		return
	}

	if tableExists("shops") && !tableExists("companies") {
		if err := db.Exec("ALTER TABLE shops RENAME TO companies").Error; err != nil {
			log.Printf("[DB] rename shops -> companies failed: %v", err)
		} else {
			log.Println("[DB] renamed table shops -> companies")
		}
	}

	if tableExists("products") && !tableExists("projects") {
		if err := db.Exec("ALTER TABLE products RENAME TO projects").Error; err != nil {
			log.Printf("[DB] rename products -> projects failed: %v", err)
		} else {
			log.Println("[DB] renamed table products -> projects")
		}
	}
}

func tableExists(name string) bool {
	var count int64
	DB().Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND table_name = ?",
		name,
	).Scan(&count)
	return count > 0
}
