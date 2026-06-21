package database

import (
	"fmt"

	"corp-site/internal/model"
)

var energySubNames = []string{
	"光伏发电", "储能电站", "风力发电", "垃圾发电", "水利发电", "充电桩", "光储充一体化项目",
}

var canonicalParentNames = []string{
	"新能源项目", "企业类项目", "政信类项目", "电站出售方", "电站收购方", "租赁公司", "其他类",
}

var legacyToNewPath = map[string]string{
	"新能源":  "新能源项目/光伏发电",
	"融资":   "租赁公司/金租",
	"租赁":   "企业类项目/中小微企业设备租赁",
	"技术合作": "其他类",
	"项目转让": "电站出售方/光伏发电",
	"其他":   "其他类",
}

type categoryGroup struct {
	name  string
	order int
	subs  []string
}

// SeedCategories upgrades legacy flat categories to the new tree on every startup when needed.
func SeedCategories() {
	db.Exec("DROP INDEX IF EXISTS idx_categories_name")

	if !hasCanonicalTree() {
		migrateFromLegacy(!hasAnyCategory())
		return
	}

	ensureCanonicalTree()

	if cleaned := cleanupLegacyCategories(); cleaned > 0 {
		fmt.Printf("[DB] removed %d legacy categories\n", cleaned)
	}

	if fixed := fixOrphanPosts(); fixed > 0 {
		fmt.Printf("[DB] remapped %d posts with invalid categories\n", fixed)
	}
}

func hasAnyCategory() bool {
	var n int64
	db.Model(&model.Category{}).Count(&n)
	return n > 0
}

func hasCanonicalTree() bool {
	var n int64
	db.Model(&model.Category{}).
		Where("name = ? AND parent_id IS NULL", "新能源项目").
		Count(&n)
	return n > 0
}

func migrateFromLegacy(freshInstall bool) {
	var oldCats []model.Category
	db.Find(&oldCats)
	oldNameByID := make(map[uint]string, len(oldCats))
	for _, c := range oldCats {
		oldNameByID[c.ID] = c.Name
	}

	var posts []model.Post
	db.Find(&posts)
	postOldCat := make(map[string]uint, len(posts))
	for _, p := range posts {
		postOldCat[p.ID.String()] = p.CategoryID
	}

	if len(oldCats) > 0 {
		db.Where("1 = 1").Delete(&model.Category{})
	}

	createCategoryTree()
	leafByPath := buildLeafPathMap()

	migrated := 0
	for _, p := range posts {
		oldID := postOldCat[p.ID.String()]
		oldName := oldNameByID[oldID]
		pathKey, ok := legacyToNewPath[oldName]
		if !ok {
			continue
		}
		newID, ok := leafByPath[pathKey]
		if !ok {
			continue
		}
		db.Model(&p).Update("category_id", newID)
		migrated++
	}

	if freshInstall {
		fmt.Println("[DB] category tree seeded")
	} else {
		fmt.Printf("[DB] category tree migrated (%d posts remapped)\n", migrated)
	}
}

func cleanupLegacyCategories() int {
	var legacy []model.Category
	db.Where("parent_id IS NULL AND name NOT IN ?", canonicalParentNames).Find(&legacy)
	if len(legacy) == 0 {
		return 0
	}

	leafByPath := buildLeafPathMap()
	legacyIDs := make([]uint, len(legacy))
	legacyNameByID := make(map[uint]string, len(legacy))
	for i, c := range legacy {
		legacyIDs[i] = c.ID
		legacyNameByID[c.ID] = c.Name
	}

	var posts []model.Post
	db.Where("category_id IN ?", legacyIDs).Find(&posts)
	for _, p := range posts {
		pathKey, ok := legacyToNewPath[legacyNameByID[p.CategoryID]]
		if !ok {
			pathKey = "其他类"
		}
		if newID, ok := leafByPath[pathKey]; ok {
			db.Model(&p).Update("category_id", newID)
		}
	}

	db.Where("id IN ?", legacyIDs).Delete(&model.Category{})
	return len(legacy)
}

func fixOrphanPosts() int {
	validIDs := buildValidCategoryIDSet()
	if len(validIDs) == 0 {
		return 0
	}

	ids := make([]uint, 0, len(validIDs))
	for id := range validIDs {
		ids = append(ids, id)
	}

	var posts []model.Post
	db.Where("category_id NOT IN ?", ids).Find(&posts)
	if len(posts) == 0 {
		return 0
	}

	leafByPath := buildLeafPathMap()
	fallbackID, ok := leafByPath["其他类"]
	if !ok {
		return 0
	}

	for _, p := range posts {
		db.Model(&p).Update("category_id", fallbackID)
	}
	return len(posts)
}

func buildValidCategoryIDSet() map[uint]struct{} {
	result := make(map[uint]struct{})
	var parents []model.Category
	db.Where("parent_id IS NULL AND name IN ?", canonicalParentNames).
		Preload("Children").Order("sort_order ASC").Find(&parents)
	for _, p := range parents {
		if len(p.Children) == 0 {
			result[p.ID] = struct{}{}
			continue
		}
		for _, ch := range p.Children {
			result[ch.ID] = struct{}{}
		}
	}
	return result
}

func canonicalGroups() []categoryGroup {
	return []categoryGroup{
		{name: "新能源项目", order: 1, subs: energySubNames},
		{name: "企业类项目", order: 2, subs: []string{"央国企设备租赁", "上市公司设备租赁", "中小微企业设备租赁"}},
		{name: "政信类项目", order: 3, subs: []string{"地级市平台公司", "区县级平台公司"}},
		{name: "电站出售方", order: 4, subs: energySubNames},
		{name: "电站收购方", order: 5, subs: energySubNames},
		{name: "租赁公司", order: 6, subs: []string{"金租", "商租", "外资"}},
		{name: "其他类", order: 7, subs: nil},
	}
}

func createCategoryTree() {
	for _, g := range canonicalGroups() {
		parent := model.Category{Name: g.name, SortOrder: g.order}
		db.Create(&parent)
		for i, sub := range g.subs {
			child := model.Category{
				ParentID:  &parent.ID,
				Name:      sub,
				SortOrder: i + 1,
			}
			db.Create(&child)
		}
	}
}

// ensureCanonicalTree adds any missing canonical parents/children and normalizes
// their sort order, so existing (already-initialized) databases pick up new
// categories without a destructive migration.
func ensureCanonicalTree() {
	for _, g := range canonicalGroups() {
		var parent model.Category
		err := db.Where("name = ? AND parent_id IS NULL", g.name).First(&parent).Error
		if err != nil {
			parent = model.Category{Name: g.name, SortOrder: g.order}
			db.Create(&parent)
			fmt.Printf("[DB] added category %s\n", g.name)
		} else if parent.SortOrder != g.order {
			db.Model(&parent).Update("sort_order", g.order)
		}
		for i, sub := range g.subs {
			var child model.Category
			if db.Where("name = ? AND parent_id = ?", sub, parent.ID).First(&child).Error != nil {
				db.Create(&model.Category{
					ParentID:  &parent.ID,
					Name:      sub,
					SortOrder: i + 1,
				})
				fmt.Printf("[DB] added category %s/%s\n", g.name, sub)
			}
		}
	}
}

func buildLeafPathMap() map[string]uint {
	result := make(map[string]uint)
	var parents []model.Category
	db.Where("parent_id IS NULL AND name IN ?", canonicalParentNames).
		Preload("Children").Order("sort_order ASC").Find(&parents)
	for _, p := range parents {
		if len(p.Children) == 0 {
			result[p.Name] = p.ID
			continue
		}
		for _, ch := range p.Children {
			result[p.Name+"/"+ch.Name] = ch.ID
		}
	}
	return result
}

// CanonicalParentNames exposes the allowed top-level navigation categories.
func CanonicalParentNames() []string {
	names := make([]string, len(canonicalParentNames))
	copy(names, canonicalParentNames)
	return names
}
