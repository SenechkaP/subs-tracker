package subscription

import (
	"context"

	"github.com/SenechkaP/subs-tracker/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var s models.Subscription
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) Update(ctx context.Context, s *models.Subscription) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error
}

func (r *Repository) ListByUser(ctx context.Context, userID *uuid.UUID) ([]models.Subscription, error) {
	var out []models.Subscription
	q := r.db.WithContext(ctx).Order("start_date desc")
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
