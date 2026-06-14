package model

import "time"

type Category struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	ParentID  *uint      `gorm:"index;uniqueIndex:idx_cat_parent_name" json:"parent_id,omitempty"`
	Name      string     `gorm:"type:varchar(50);not null;uniqueIndex:idx_cat_parent_name" json:"name"`
	SortOrder int        `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	Parent    *Category  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children  []Category `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

func (c Category) FullName() string {
	if c.Parent != nil {
		return c.Parent.Name + " · " + c.Name
	}
	return c.Name
}
