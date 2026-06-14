package handler

import (
	"corp-site/internal/database"
	"corp-site/internal/model"
)

type CategoryChildStat struct {
	ID           uint
	Name         string
	PostCount    int64
	ProductCount int64
}

type CategoryParentStat struct {
	ID           uint
	Name         string
	PostCount    int64
	ProductCount int64
	Children     []CategoryChildStat
}

func LoadCategoryStats() []CategoryParentStat {
	db := database.DB()

	postCounts := map[uint]int64{}
	var postRows []struct {
		CategoryID uint
		Count      int64
	}
	db.Model(&model.Post{}).Select("category_id, COUNT(*) as count").Group("category_id").Scan(&postRows)
	for _, r := range postRows {
		postCounts[r.CategoryID] = r.Count
	}

	productCounts := map[uint]int64{}
	var productRows []struct {
		CategoryID uint
		Count      int64
	}
	db.Model(&model.Project{}).Select("category_id, COUNT(*) as count").Group("category_id").Scan(&productRows)
	for _, r := range productRows {
		productCounts[r.CategoryID] = r.Count
	}

	nav := LoadCategoryNav()
	stats := make([]CategoryParentStat, len(nav))
	for i, parent := range nav {
		ps := CategoryParentStat{
			ID:           parent.ID,
			Name:         parent.Name,
			PostCount:    postCounts[parent.ID],
			ProductCount: productCounts[parent.ID],
		}
		if len(parent.Children) == 0 {
			ps.Children = []CategoryChildStat{}
		} else {
			ps.Children = make([]CategoryChildStat, len(parent.Children))
			for j, ch := range parent.Children {
				ps.Children[j] = CategoryChildStat{
					ID:           ch.ID,
					Name:         ch.Name,
					PostCount:    postCounts[ch.ID],
					ProductCount: productCounts[ch.ID],
				}
				ps.PostCount += postCounts[ch.ID]
				ps.ProductCount += productCounts[ch.ID]
			}
		}
		stats[i] = ps
	}
	return stats
}
