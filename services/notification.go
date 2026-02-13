package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/HuseyinAsik/Notifications/models"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/repository"
	"github.com/HuseyinAsik/Notifications/serializers"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	pageLimit = 20
)

type NotificationService struct {
	NotificationRepo repository.NotificationRepository
	Logger           *logging.LogWrapper
}

func NewNotificationService(notificationRepo repository.NotificationRepository, logger *logging.LogWrapper) *NotificationService {
	return &NotificationService{
		NotificationRepo: notificationRepo,
		Logger:           logger}
}

func (s *NotificationService) Create(
	ctx context.Context,
	form *serializers.CreateNotificationForm,
) (string, error) {
	Id := uuid.NewString()
	notification := models.Notification{
		Id:          Id,
		GroupId:     Id,
		Recipient:   form.Recipient,
		Channel:     form.Channel,
		Content:     form.Content,
		Status:      "pending",
		Priority:    form.Priority,
		ScheduledAt: form.ScheduledAt,
	}
	event := CreateEvent(notification)
	err := s.NotificationRepo.Create(ctx, notification, event)

	if err != nil {
		s.Logger.Error(ctx, "Notification Create Err", zap.Error(err))
		return "", err
	}

	return Id, nil
}

func (s *NotificationService) BulkCreate(ctx context.Context, batchForm serializers.CreateNotificationBatchForm) (string, error) {
	var notifications []models.Notification
	var events []*models.OutboxEvent

	groupId := uuid.NewString()
	now := time.Now()
	for _, data := range batchForm.Data {
		notification := models.Notification{
			Id:          uuid.NewString(),
			GroupId:     groupId,
			Recipient:   data.Recipient,
			Channel:     data.Channel,
			Content:     data.Content,
			Status:      "pending",
			Priority:    data.Priority,
			ScheduledAt: data.ScheduledAt,
			CreatedAt:   now,
		}
		notifications = append(notifications, notification)

		event := CreateEvent(notification)
		if event != nil {
			events = append(events, event)
		}
	}

	err := s.NotificationRepo.BulkInsertWithOutbox(ctx, notifications, events)

	if err != nil {
		s.Logger.Error(ctx, "Notification service BulkInsertWithOutbox err", zap.Error(err))
		return "", err
	}

	return groupId, nil
}

func (s *NotificationService) List(ctx context.Context, listForm serializers.ListForm) ([]models.Notification, int, error) {
	offset := (listForm.Page - 1) * pageLimit

	notifications, total, err := s.NotificationRepo.ListNotifications(ctx, listForm.Status, listForm.Channel, listForm.StartDate, listForm.EndDate, pageLimit, offset)

	if err != nil {
		s.Logger.Error(ctx, "Notification List Err", zap.Error(err))
	}

	return notifications, total, err
}

func BuildTopic(notificationType, priority string) string {
	return fmt.Sprintf("%s_%s", notificationType, priority)
}

func CreateEvent(notification models.Notification) *models.OutboxEvent {
	now := time.Now()
	if notification.ScheduledAt != nil && notification.ScheduledAt.After(now) {
		return nil
	}

	topic := BuildTopic(notification.Channel, notification.Priority)

	payload, _ := json.Marshal(notification)

	event := &models.OutboxEvent{
		Id:          uuid.NewString(),
		AggregateId: notification.Id,
		GroupId:     notification.GroupId,
		EventType:   "NotificationCreated",
		Topic:       topic,
		Payload:     payload,
		CreatedAt:   now,
	}

	return event
}
