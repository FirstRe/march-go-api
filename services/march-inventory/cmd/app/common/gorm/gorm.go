package gormDb

import (
	"march-inventory/cmd/app/graph/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Repos *gorm.DB
var InventoryType *gorm.DB
var InventoryBrand *gorm.DB

func Initialize() (*gorm.DB, error) {
	dsn := "root:123456@tcp(0.0.0.0:3306)/march_inventory_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})

	if err != nil {
		return nil, err
	}
	// db.Callback().Create().Before("gorm:before_create").Register("custom_before_create", BeforeCreate)
	InventoryType = db.Model(&model.InventoryType{})
	Repos = db
	return db, nil
}
