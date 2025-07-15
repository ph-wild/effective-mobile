package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"effective_mobile/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type SubscriptionStorage struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewSubscriptionStorage(db *sqlx.DB, log *slog.Logger) *SubscriptionStorage {
	return &SubscriptionStorage{
		db:  db,
		log: log,
	}
}

func (s *SubscriptionStorage) Create(ctx context.Context, sub *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (service_name, price, user_id, start_date, end_date)
		VALUES (:service_name, :price, :user_id, :start_date, :end_date)
		RETURNING id, created_at, updated_at
	`

	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		s.log.Error("prepare insert failed", slog.Any("err", err))
		return errors.Wrap(err, "prepare insert")
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, sub, sub); err != nil {
		s.log.Error("insert failed", slog.Any("err", err))
		return errors.Wrap(err, "insert subscription")
	}

	return nil
}

func (s *SubscriptionStorage) GetByID(ctx context.Context, id int) (*models.Subscription, error) {
	var sub models.Subscription
	query := `SELECT * FROM subscriptions WHERE id = $1`

	if err := s.db.GetContext(ctx, &sub, query, id); err != nil {
		return nil, errors.Wrap(err, "get by id")
	}

	return &sub, nil
}

func (s *SubscriptionStorage) List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]models.Subscription, error) {
	query := `SELECT * FROM subscriptions WHERE 1=1`
	args := []interface{}{}
	i := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", i)
		args = append(args, *userID)
		i++
	}
	if serviceName != nil {
		query += fmt.Sprintf(" AND service_name ILIKE $%d", i)
		args = append(args, "%"+*serviceName+"%")
		i++
	}
	query += fmt.Sprintf(" ORDER BY id LIMIT $%d OFFSET $%d", i, i+1)
	args = append(args, limit, offset)

	var subs []models.Subscription
	if err := s.db.SelectContext(ctx, &subs, query, args...); err != nil {
		return nil, errors.Wrap(err, "list subscriptions")
	}
	return subs, nil
}

func (s *SubscriptionStorage) Update(ctx context.Context, id int, sub *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET service_name = :service_name,
		    price = :price,
		    user_id = :user_id,
		    start_date = :start_date,
		    end_date = :end_date,
		    updated_at = now()
		WHERE id = :id
	`

	sub.ID = id
	_, err := s.db.NamedExecContext(ctx, query, sub)
	if err != nil {
		return errors.Wrap(err, "update subscription")
	}
	return nil
}

func (s *SubscriptionStorage) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM subscriptions WHERE id = $1`
	if _, err := s.db.ExecContext(ctx, query, id); err != nil {
		return errors.Wrap(err, "delete subscription")
	}
	return nil
}

func (s *SubscriptionStorage) SumByPeriod(ctx context.Context, userID uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE user_id = $1
		  AND start_date >= $2
		  AND (end_date IS NULL OR end_date <= $3)
	`
	args := []interface{}{userID, from, to}

	if serviceName != nil {
		query += " AND service_name ILIKE $4"
		args = append(args, "%"+*serviceName+"%")
	}

	var total int
	if err := s.db.GetContext(ctx, &total, query, args...); err != nil {
		return 0, errors.Wrap(err, "sum by period")
	}
	return total, nil
}
