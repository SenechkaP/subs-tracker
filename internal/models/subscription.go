package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Subscription struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;index" json:"id"`
	Service   string     `gorm:"not null" json:"service_name"`
	PriceRUB  int64      `gorm:"not null" json:"price"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	StartDate time.Time  `gorm:"not null" json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (s *Subscription) GenerateNewUUID(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
