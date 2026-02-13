package services

import (
	"context"
	"time"

	"github.com/HuseyinAsik/Notifications/pkg/kafka"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/repository"
	"go.uber.org/zap"
)

type Outbox struct {
	repo      repository.NotificationRepository
	writer    *kafka.Writer
	logger    *logging.LogWrapper
	batchSize int
}

func NewOutbox(
	repo repository.NotificationRepository,
	writer *kafka.Writer,
	logger *logging.LogWrapper,
) *Outbox {
	return &Outbox{
		repo:      repo,
		writer:    writer,
		logger:    logger,
		batchSize: 100,
	}
}

func (p *Outbox) Run(ctx context.Context) {

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.process(ctx)
		}
	}
}

func (p *Outbox) process(ctx context.Context) {

	events, err := p.repo.FetchPendingOutbox(ctx, p.batchSize)
	if err != nil {
		p.logger.Error(ctx, "Outbox process Err", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	var messages []kafka.Message
	var ids []string

	for _, e := range events {
		messages = append(messages, kafka.Message{
			Topic: e.Topic,
			Value: e.Payload,
		})
		ids = append(ids, e.Id)
	}

	if markPublishedErr := p.repo.MarkOutboxPublished(ctx, ids); markPublishedErr != nil {
		p.logger.Error(ctx, "Outbox MarkOutboxPublished Err", zap.Error(markPublishedErr))
		return
	}

	if writeMessageErr := p.writer.WriteMessages(ctx, messages); writeMessageErr != nil {
		p.logger.Error(ctx, "Outbox WriteMessages Err", zap.Error(writeMessageErr))

		if markpendingErr := p.repo.MarkOutboxPending(ctx, ids); markpendingErr != nil {
			p.logger.Error(ctx, "Outbox MarkOutboxPending Err", zap.Error(markpendingErr), zap.Strings("ids", ids))
		}
		return
	}

}
