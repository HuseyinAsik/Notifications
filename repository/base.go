package repository

import (
	"context"
	"time"

	"github.com/HuseyinAsik/Notifications/models"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification models.Notification, event *models.OutboxEvent) error
	BulkInsertWithOutbox(ctx context.Context, notifications []models.Notification, events []*models.OutboxEvent) error
	FetchPendingOutbox(ctx context.Context, limit int) ([]models.OutboxEvent, error)
	FetchOutboxEventByAggregateId(ctx context.Context, Id string) (*models.OutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, ids []string) error
	MarkOutboxPending(ctx context.Context, ids []string) error
	UpdateOutboxEvent(ctx context.Context, Id, status string, retryCount int) error
	UpdateNotificationStatus(ctx context.Context, Id, status string) error
	ListNotifications(ctx context.Context, status, channel string, startDate, endDate *time.Time, limit, offset int) ([]models.Notification, int, error)
	FindById(ctx context.Context, id string) (*models.Notification, error)
}
