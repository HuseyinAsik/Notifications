package worker

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/HuseyinAsik/Notifications/models"
	"github.com/HuseyinAsik/Notifications/providers"
	"github.com/HuseyinAsik/Notifications/repository"

	gkafka "github.com/HuseyinAsik/Notifications/pkg/kafka"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
)

const batchSize = 100

type Worker struct {
	highReader   *kafka.Reader
	normalReader *kafka.Reader
	lowReader    *kafka.Reader

	limiter  *rate.Limiter
	provider providers.Provider
	repo     repository.NotificationRepository
	logger   *logging.LogWrapper
}
type FetchedMessage struct {
	Reader  *kafka.Reader
	Message kafka.Message
}

func NewWorker(
	brokers []string,
	channel string,
	rateLimit int,
	prov providers.Provider,
	repo repository.NotificationRepository,
	logger *logging.LogWrapper,
) *Worker {

	groupID := channel + "-worker-group"

	return &Worker{
		highReader:   gkafka.NewReader(brokers, channel+"_high", groupID),
		normalReader: gkafka.NewReader(brokers, channel+"_medium", groupID),
		lowReader:    gkafka.NewReader(brokers, channel+"_low", groupID),
		limiter:      rate.NewLimiter(rate.Limit(rateLimit), rateLimit),
		provider:     prov,
		repo:         repo,
		logger:       logger,
	}
}

func (w *Worker) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			w.shutdown()
			return nil
		default:
			w.processBatch(ctx)
		}
	}
}
func (w *Worker) processBatch(ctx context.Context) {
	messages := make([]FetchedMessage, 0, batchSize)

	messages = append(messages, w.fetch(ctx, w.highReader, batchSize)...)

	if len(messages) < batchSize {
		messages = append(messages, w.fetch(ctx, w.normalReader, batchSize-len(messages))...)
	}

	if len(messages) < batchSize {
		messages = append(messages, w.fetch(ctx, w.lowReader, batchSize-len(messages))...)
	}

	if len(messages) == 0 {
		time.Sleep(50 * time.Millisecond)
		return
	}

	w.handle(messages, ctx)
}
func (w *Worker) fetch(ctx context.Context, reader *kafka.Reader, limit int) []FetchedMessage {
	var msgs []FetchedMessage

	for i := 0; i < limit; i++ {
		ctxFetch, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		msg, err := reader.FetchMessage(ctxFetch)
		cancel()

		if err != nil {
			break
		}

		msgs = append(msgs, FetchedMessage{
			Reader:  reader,
			Message: msg,
		})
	}

	return msgs
}
func (w *Worker) handle(messages []FetchedMessage, ctx context.Context) {
	var wg sync.WaitGroup

	for _, msg := range messages {
		wg.Add(1)

		go func(m FetchedMessage) {
			defer wg.Done()

			var n models.Notification
			if err := json.Unmarshal(m.Message.Value, &n); err != nil {
				return
			}

			w.limiter.Wait(ctx)
			ok := w.CheckEvent(ctx, n.Id)
			if ok {
				if markErr := w.MarkEvent(ctx, n.Id, true); markErr != nil {
					w.logger.Error(ctx, "handle markevent err", zap.Error(markErr))
					return
				}
				if updateNotificationErr := w.UpdateNotification(ctx, n.Id, "processing"); updateNotificationErr != nil {
					w.logger.Error(ctx, "handle updateNotification err",
						zap.Error(updateNotificationErr),
						zap.String("id", n.Id),
						zap.String("status", "processing"))
					return
				}
				if sendErr := w.provider.Send(n.Id, n.Recipient, n.Content); sendErr != nil {
					w.logger.Error(ctx, "handle send err", zap.Error(sendErr))
					if markErr := w.MarkEvent(ctx, n.Id, false); markErr != nil {
						w.logger.Error(ctx, "handle markevent err", zap.Error(markErr))
						return
					}
					return
				}
			}

			if updateNotificationErr := w.UpdateNotification(ctx, n.Id, "sended"); updateNotificationErr != nil {
				w.logger.Error(ctx, "handle updateNotification err",
					zap.Error(updateNotificationErr),
					zap.String("id", n.Id),
					zap.String("status", "sended"))
				return
			}
			w.commit(ctx, m.Message, m.Reader)
		}(msg)
	}

	wg.Wait()
}
func (w *Worker) commit(ctx context.Context, msg kafka.Message, reader *kafka.Reader) error {
	return reader.CommitMessages(ctx, msg)
}
func (w *Worker) shutdown() {
	w.highReader.Close()
	w.normalReader.Close()
	w.lowReader.Close()
}

func (w *Worker) CheckEvent(ctx context.Context, id string) bool {
	event, err := w.repo.FetchOutboxEventByAggregateId(ctx, id)

	if err != nil {
		w.logger.Error(ctx, "Worker checkevent err", zap.Error(err))
		return false
	}
	if event == nil {
		w.logger.Error(ctx, "Worker checkevent err", zap.Error(errors.New("event not found")), zap.String("id", id))
		return false
	}
	if !strings.EqualFold(event.Status, "published") || event.RetryCount > 6 {
		return false
	}
	return true
}

func (w *Worker) MarkEvent(ctx context.Context, id string, result bool) error {
	event, err := w.repo.FetchOutboxEventByAggregateId(ctx, id)
	status := "sended"
	tryCount := event.RetryCount + 1
	if err != nil {
		return err
	}

	if !result && event.RetryCount < 6 {
		status = "pending"
	}

	if !result && event.RetryCount >= 6 {
		status = "failed"
	}
	updateErr := w.repo.UpdateOutboxEvent(ctx, id, status, tryCount)

	if updateErr != nil {
		return updateErr
	}

	return nil
}

func (w *Worker) UpdateNotification(ctx context.Context, id, status string) error {
	err := w.repo.UpdateNotificationStatus(ctx, id, status)

	return err
}
