package basedb

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Base) BeforeCreate(scope *gorm.DB) (err error) {
	uuid := uuid.New()
	base.ID = uuid
	return
}

type UserRe struct {
	Base
	SomeFlag bool `gorm:"column:some_flag;not null;default:true"`
}
