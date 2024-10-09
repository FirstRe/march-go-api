package gormDb

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Repos *gorm.DB

func Initialize() (*gorm.DB, error) {
	dsn := "root:123456@tcp(0.0.0.0:3306)/march-user-test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	// db.Callback().Create().Before("gorm:before_create").Register("custom_before_create", BeforeCreate)
	
	Repos = db
	return db, nil
}
