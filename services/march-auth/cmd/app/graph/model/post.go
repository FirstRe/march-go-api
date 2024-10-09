package model

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	ID          string `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Title       string `json:"title"`
	Label       string `json:"label"`
	UserID      string `gorm:"column:userId" json:"userId"`
	User        User   `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Description string `json:"description"`
	CreatedAt   string `gorm:"type:datetime;default:now()" json:"createdAt"`
	UpdatedAt   string `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (base *Post) BeforeCreate(scope *gorm.DB) (err error) {
	uuid := uuid.New()
	base.ID = uuid.String()
	return
}
