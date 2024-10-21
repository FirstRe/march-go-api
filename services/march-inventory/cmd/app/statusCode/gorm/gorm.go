package gormDb

import (
	"march-inventory/cmd/app/graph/model"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Repos *gorm.DB
var InventoryType *gorm.DB
var InventoryBrand *gorm.DB

func Initialize() (*gorm.DB, error) {
	dsn := viper.GetString("DATABASE_URL")
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
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
