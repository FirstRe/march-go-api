package gormDb

import (
	"march-inventory/cmd/app/graph/model"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var Repos *gorm.DB
var InventoryType *gorm.DB
var InventoryBrand *gorm.DB

func Initialize() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		// PrepareStmt:            true,
		PrepareStmt: false,
		// SkipDefaultTransaction: true,
		// TranslateError:         true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: "march_inventory.",
		},
	})
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
	// 	Logger:                 logger.Default.LogMode(logger.Info),
	// 	// PrepareStmt:            true,
	// 	// SkipDefaultTransaction: true,
	// 	// TranslateError:         true,
	// 	NamingStrategy: schema.NamingStrategy{
	// 		TablePrefix: "march_inventory.",
	// 	},
	// })

	if err != nil {
		return nil, err
	}
	// db.Callback().Create().Before("gorm:before_create").Register("custom_before_create", BeforeCreate)
	InventoryType = db.Model(&model.InventoryType{})
	Repos = db
	return db, nil
}
