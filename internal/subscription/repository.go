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

func (repository *SubscriptionRepository) Create(ctx context.Context, s *models.Subscription) error {
	return repository.db.WithContext(ctx).Create(s).Error
}

func (repository *SubscriptionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Subscription, error) {
	var s models.Subscription
	if err := repository.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (repository *SubscriptionRepository) Update(ctx context.Context, s *models.Subscription) (*models.Subscription, error) {
	if err := repository.db.WithContext(ctx).Save(s).Error; err != nil {
		return nil, err
	}
	return s, nil
}

func (repository *SubscriptionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := repository.db.WithContext(ctx).Delete(&models.Subscription{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (repository *SubscriptionRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]models.Subscription, error) {
	var out []models.Subscription
	q := repository.db.WithContext(ctx).Where("user_id = ?", userID).Order("start_date desc")
	if err := q.Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
