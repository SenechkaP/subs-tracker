package db

import (
	"fmt"
	"log"

	"github.com/SenechkaP/subs-tracker/configs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDb(conf *configs.Config) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		conf.DBHost, conf.DBUser, conf.DBPassword, conf.DBName, conf.DBPort)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	return gormDB
}
