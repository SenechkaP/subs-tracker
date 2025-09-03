package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/SenechkaP/subs-tracker/internal/models"
)

func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "20250903_create_subscriptions",
			Migrate: func(tx *gorm.DB) error {
				return tx.AutoMigrate(&models.Subscription{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("subscriptions")
			},
		},
	}
}

func RunMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, GetMigrations())
	return m.Migrate()
}
