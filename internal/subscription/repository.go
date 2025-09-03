package subscription

import (
	"context"

	"github.com/SenechkaP/subs-tracker/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (handler *SubscriptionRepository) Create(ctx context.Context, s *models.Subscription) error {
	return handler.db.WithContext(ctx).Create(s).Error
}

func (handler *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var s models.Subscription
	if err := handler.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (handler *SubscriptionRepository) Update(ctx context.Context, s *models.Subscription) error {
	return handler.db.WithContext(ctx).Save(s).Error
}

func (handler *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return handler.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id).Error
}

func (handler *SubscriptionRepository) ListByUser(ctx context.Context, userID *uuid.UUID) ([]models.Subscription, error) {
	var out []models.Subscription
	q := handler.db.WithContext(ctx).Order("start_date desc")
	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
