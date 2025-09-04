package subscription

import (
	"context"
	"database/sql"
	"time"

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

func (repo *SubscriptionRepository) SumPriceByMonthRange(
	ctx context.Context,
	intervalStart time.Time,
	intervalEnd time.Time,
	userID *uuid.UUID,
	service *string,
) (int64, error) {
	var total sql.NullInt64

	q := repo.db.WithContext(ctx).
		Model(&models.Subscription{}).
		Select("COALESCE(SUM(price_rub), 0) as total").
		Where(`
            (
              (end_date IS NOT NULL AND start_date <= ? AND end_date >= ?)
              OR
              (end_date IS NULL AND start_date BETWEEN ? AND ?)
            )`, intervalEnd, intervalStart, intervalStart, intervalEnd)

	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}
	if service != nil && *service != "" {
		q = q.Where("service = ?", *service)
	}

	if err := q.Scan(&total).Error; err != nil {
		return 0, err
	}
	if !total.Valid {
		return 0, nil
	}
	return total.Int64, nil
}
