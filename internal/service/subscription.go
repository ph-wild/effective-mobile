package service

import (
	"context"
	"log/slog"
	"time"

	"effective_mobile/internal/models"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type SubscriptionServiceInterface interface {
	Create(ctx context.Context, sub *models.Subscription) error
	GetByID(ctx context.Context, id int) (*models.Subscription, error)
	List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error)
	Update(ctx context.Context, id int, sub *models.Subscription) error
	Delete(ctx context.Context, id int) error
	SumByPeriod(ctx context.Context, userID uuid.UUID, serviceName *string, from, to time.Time) (int, error)
}

type SubscriptionService struct {
	storage SubscriptionServiceInterface
	log     *slog.Logger
}

func NewSubscriptionService(s SubscriptionServiceInterface, log *slog.Logger) *SubscriptionService {
	return &SubscriptionService{
		storage: s,
		log:     log,
	}
}

func (s *SubscriptionService) Create(ctx context.Context, sub *models.Subscription) error {
	if sub.Price < 0 {
		return errors.New("price must be non-negative")
	}

	if err := s.storage.Create(ctx, sub); err != nil {
		s.log.Error("storage create failed", slog.Any("err", err))
		return errors.Wrap(err, "service create")
	}

	return nil
}

func (s *SubscriptionService) Get(ctx context.Context, id int) (*models.Subscription, error) {
	sub, err := s.storage.GetByID(ctx, id)
	if err != nil {
		s.log.Error("get by id failed", slog.Int("id", id), slog.Any("err", err))
		return nil, errors.Wrap(err, "service get")
	}
	return sub, nil
}

func (s *SubscriptionService) List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	return s.storage.List(ctx, userID, serviceName, limit, offset)
}

func (s *SubscriptionService) Update(ctx context.Context, id int, sub *models.Subscription) error {
	sub.ID = id
	return s.storage.Update(ctx, id, sub)
}

func (s *SubscriptionService) Delete(ctx context.Context, id int) error {
	return s.storage.Delete(ctx, id)
}

func (s *SubscriptionService) SumByPeriod(ctx context.Context, userID uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	return s.storage.SumByPeriod(ctx, userID, serviceName, from, to)
}
