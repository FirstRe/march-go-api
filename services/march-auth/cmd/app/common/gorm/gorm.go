package gormDb

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Repos *gorm.DB

func Initialize() (*gorm.DB, error) {
	dsn := viper.GetString("DATABASE_URL")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	// db.Callback().Create().Before("gorm:before_create").Register("custom_before_create", BeforeCreate)

	Repos = db
	return db, nil
}
