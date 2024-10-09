package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        string    `gorm:"type:varchar(40);primaryKey;default:uuid()" json:"id"`
	Name      string    `gorm:"not null;unique" json:"name"`
	Email     string    `gorm:"not null;unique" json:"email"`
	Password  string    `gorm:"not null" json:"password"`
	CreatedAt string    `gorm:"type:datetime;default:now()" json:"createdAt"`
	UpdatedAt time.Time `gorm:"type:datetime;autoUpdateTime" json:"updatedAt"`
	// Posts     []Post `gorm:"foreignKey:UserID;references:ID" json:"posts"`
}

func (base *User) BeforeCreate(scope *gorm.DB) (err error) {
	uuid := uuid.New()
	base.ID = uuid.String()
	return
}
