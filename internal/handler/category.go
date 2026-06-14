package handler

import (
	"fmt"
	"strconv"

	"corp-site/internal/database"
	"corp-site/internal/model"
)

type CategoryNavChild struct {
	ID   uint
	Name string
}

type CategoryNavItem struct {
	ID       uint
	Name     string
	Children []CategoryNavChild
}

func LoadCategoryNav() []CategoryNavItem {
	var parents []model.Category
	database.DB().Where("parent_id IS NULL AND name IN ?", database.CanonicalParentNames()).
		Order("sort_order ASC").Find(&parents)

	items := make([]CategoryNavItem, len(parents))
	for i, p := range parents {
		item := CategoryNavItem{ID: p.ID, Name: p.Name}
		var children []model.Category
		database.DB().Where("parent_id = ?", p.ID).Order("sort_order ASC").Find(&children)
		for _, ch := range children {
			item.Children = append(item.Children, CategoryNavChild{ID: ch.ID, Name: ch.Name})
		}
		items[i] = item
	}
	return items
}

func LoadLeafCategories() []model.Category {
	var parents []model.Category
	database.DB().Where("parent_id IS NULL AND name IN ?", database.CanonicalParentNames()).
		Order("sort_order ASC").Find(&parents)

	var leaves []model.Category
	for _, p := range parents {
		var children []model.Category
		database.DB().Where("parent_id = ?", p.ID).Order("sort_order ASC").Find(&children)
		if len(children) == 0 {
			leaves = append(leaves, p)
			continue
		}
		for _, ch := range children {
			parentCopy := p
			ch.Parent = &parentCopy
			leaves = append(leaves, ch)
		}
	}
	return leaves
}

func resolveActiveCategory(categoryIDStr string) (activeCategoryID, activeParentID, activeCategoryName string) {
	if categoryIDStr == "" {
		return "", "", ""
	}
	var cat model.Category
	if err := database.DB().Preload("Parent").First(&cat, categoryIDStr).Error; err != nil {
		return categoryIDStr, "", ""
	}
	activeCategoryID = categoryIDStr
	if cat.ParentID != nil && cat.Parent != nil {
		activeParentID = strconv.FormatUint(uint64(*cat.ParentID), 10)
		activeCategoryName = cat.Parent.Name + " · " + cat.Name
		return activeCategoryID, activeParentID, activeCategoryName
	}
	activeParentID = strconv.FormatUint(uint64(cat.ID), 10)
	activeCategoryName = cat.Name
	return activeCategoryID, activeParentID, activeCategoryName
}

func formatPostCategory(cat model.Category) string {
	if cat.Parent != nil {
		return cat.Parent.Name + " · " + cat.Name
	}
	if cat.ParentID != nil {
		var parent model.Category
		if database.DB().First(&parent, *cat.ParentID).Error == nil {
			return parent.Name + " · " + cat.Name
		}
	}
	return cat.Name
}

func catColorClass(parentName string) string {
	switch parentName {
	case "新能源项目":
		return "bg-emerald-100 text-emerald-700"
	case "企业类项目":
		return "bg-sky-100 text-sky-700"
	case "电站出售方":
		return "bg-orange-100 text-orange-700"
	case "电站收购方":
		return "bg-violet-100 text-violet-700"
	case "租赁公司":
		return "bg-amber-100 text-amber-700"
	case "其他类":
		return "bg-gray-100 text-gray-600"
	default:
		return "bg-blue-100 text-blue-700"
	}
}

func catParentName(cat model.Category) string {
	if cat.Parent != nil {
		return cat.Parent.Name
	}
	if cat.ParentID != nil {
		var parent model.Category
		if database.DB().First(&parent, *cat.ParentID).Error == nil {
			return parent.Name
		}
	}
	return cat.Name
}

func uintStr(v uint) string {
	return fmt.Sprintf("%d", v)
}
