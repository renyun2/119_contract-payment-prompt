package database

import (
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=127.0.0.1 user=contractpay password=contractpay dbname=contract_payment port=5432 sslmode=disable TimeZone=Asia/Shanghai"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().In(time.FixedZone("CST", 8*3600))
		},
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(4)
	sqlDB.SetMaxOpenConns(32)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}
