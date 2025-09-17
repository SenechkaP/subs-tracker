package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func GetMigrations() []*gormigrate.Migration {
	return []*gormigrate.Migration{
		{
			ID: "20250917_create_subscriptions",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec(`
					CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

					CREATE TABLE IF NOT EXISTS subscriptions (
						id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
						service VARCHAR(255) NOT NULL,
						price_rub BIGINT NOT NULL,
						user_id UUID NOT NULL,
						start_date TIMESTAMP NOT NULL,
						end_date TIMESTAMP NULL,
						created_at TIMESTAMP NOT NULL DEFAULT NOW(),
						updated_at TIMESTAMP NOT NULL DEFAULT NOW()
					);

					CREATE INDEX idx_subscriptions_id ON subscriptions(id);
					CREATE INDEX idx_subscriptions_user_id ON subscriptions(user_id);
				`).Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec(`
					DROP INDEX IF EXISTS idx_subscriptions_id;
					DROP INDEX IF EXISTS idx_subscriptions_user_id;
					DROP TABLE IF EXISTS subscriptions;
				`).Error
			},
		},
	}
}

func RunMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, GetMigrations())
	return m.Migrate()
}
