package gormDb

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

var Repos *gorm.DB

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
			TablePrefix: "march_auth.",
		},
	})
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
	// 	Logger: logger.Default.LogMode(logger.Info),
	// 	// PrepareStmt:            true,
	// 	PrepareStmt:          false,
	// 	PreferSimpleProtocol: true,
	// 	// SkipDefaultTransaction: true,
	// 	// TranslateError:         true,
	// 	NamingStrategy: schema.NamingStrategy{
	// 		TablePrefix: "march_auth.",
	// 	},
	// })

	if err != nil {
		return nil, err
	}
	// db.Callback().Create().Before("gorm:before_create").Register("custom_before_create", BeforeCreate)

	Repos = db
	return db, nil
}
