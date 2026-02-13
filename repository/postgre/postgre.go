package postgre

import (
	"context"
	"strconv"
	"time"

	"github.com/HuseyinAsik/Notifications/models"
	"github.com/HuseyinAsik/Notifications/pkg/gpostgresql"
	"github.com/jackc/pgx/v5"
)

type PostgresNotificationRepository struct {
	db *gpostgresql.Pool
}

func NewPostgresNotificationRepository(db *gpostgresql.Pool) *PostgresNotificationRepository {
	return &PostgresNotificationRepository{db: db}
}

func (r *PostgresNotificationRepository) Create(ctx context.Context, notification models.Notification, event *models.OutboxEvent) error {

	tx, err := r.db.Write.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO notifications (
			id,
			group_id,
			recipient,
			channel,
			content,
			status,
			priority,
			scheduled_at,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`,
		notification.Id,
		notification.GroupId,
		notification.Recipient,
		notification.Channel,
		notification.Content,
		notification.Status,
		notification.Priority,
		notification.ScheduledAt,
	)

	if event != nil {
		_, err = tx.Exec(ctx, `
		INSERT INTO outbox (
			id,
			aggregate_id,
			group_id,
			event_type,
			topic,
			payload,
			status,
			retry_count,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', 0, NOW())
	`,
			event.Id,
			event.AggregateId,
			event.GroupId,
			event.EventType,
			event.Topic,
			event.Payload,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresNotificationRepository) BulkInsertWithOutbox(ctx context.Context, notifications []models.Notification, events []*models.OutboxEvent) error {

	tx, err := r.db.Write.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := copyNotifications(ctx, tx, notifications); err != nil {
		return err
	}

	if len(events) > 0 {
		if err := copyOutbox(ctx, tx, events); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresNotificationRepository) FetchPendingOutbox(ctx context.Context, limit int) ([]models.OutboxEvent, error) {

	rows, err := r.db.Read.Query(ctx, `
		SELECT id, aggregate_id, event_type,
		       topic, payload, retry_count, created_at
		FROM outbox
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.OutboxEvent

	for rows.Next() {
		var e models.OutboxEvent
		err := rows.Scan(
			&e.Id,
			&e.AggregateId,
			&e.EventType,
			&e.Topic,
			&e.Payload,
			&e.RetryCount,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, nil
}

func (r *PostgresNotificationRepository) MarkOutboxPublished(ctx context.Context, ids []string) error {

	_, err := r.db.Write.Exec(ctx, `
		UPDATE outbox
		SET status = 'published',
		    published_at = now()
		WHERE id = ANY($1)
	`, ids)

	return err
}

func (r *PostgresNotificationRepository) MarkOutboxPending(ctx context.Context, ids []string) error {

	_, err := r.db.Write.Exec(ctx, `
		UPDATE outbox
		SET status = 'pending'
		WHERE id = ANY($1)
	`, ids)

	return err
}

func (r *PostgresNotificationRepository) FetchOutboxEventByAggregateId(ctx context.Context, Id string) (*models.OutboxEvent, error) {
	query := `
	SELECT id, aggregate_id, status, retry_count
	FROM outbox
	WHERE aggregate_id = $1
`

	row := r.db.Read.QueryRow(ctx, query, Id)

	var n models.OutboxEvent
	if err := row.Scan(
		&n.Id,
		&n.AggregateId,
		&n.Status,
		&n.RetryCount,
	); err != nil {
		return nil, err
	}

	return &n, nil
}

func (r *PostgresNotificationRepository) UpdateOutboxEvent(ctx context.Context, Id, status string, retryCount int) error {
	_, err := r.db.Write.Exec(ctx, `
    UPDATE outbox
    SET
        status = $1,
        retry_count = $2
    WHERE aggregate_id = $3
`, status, retryCount, Id)

	return err
}

func (r *PostgresNotificationRepository) UpdateNotificationStatus(ctx context.Context, Id, status string) error {
	_, err := r.db.Write.Exec(ctx, `
    UPDATE notifications
    SET
        status = $1
    WHERE id = $2
`, status, Id)

	return err
}

func (r *PostgresNotificationRepository) ListNotifications(ctx context.Context, status, channel string, startDate, endDate *time.Time, limit, offset int) ([]models.Notification, int, error) {
	notifications := []models.Notification{}
	args := []interface{}{}
	where := "WHERE 1=1"

	if status != "" {
		args = append(args, status)
		where += " AND status = $" + strconv.Itoa(len(args))
	}
	if channel != "" {
		args = append(args, channel)
		where += " AND channel = $" + strconv.Itoa(len(args))
	}
	if startDate != nil {
		args = append(args, *startDate)
		where += " AND created_at >= $" + strconv.Itoa(len(args))
	}
	if endDate != nil {
		args = append(args, *endDate)
		where += " AND created_at <= $" + strconv.Itoa(len(args))
	}

	// Total count
	totalQuery := "SELECT COUNT(*) FROM notifications " + where
	var total int
	err := r.db.Read.QueryRow(ctx, totalQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	args = append(args, limit, offset)
	query := "SELECT id, recipient, channel, priority, content, status, scheduled_at, created_at FROM notifications " +
		where + " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(len(args)-1) + " OFFSET $" + strconv.Itoa(len(args))

	rows, err := r.db.Read.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var n models.Notification
		err := rows.Scan(
			&n.Id, &n.Recipient, &n.Channel, &n.Priority,
			&n.Content, &n.Status, &n.ScheduledAt, &n.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		notifications = append(notifications, n)
	}

	return notifications, total, nil
}

func copyNotifications(ctx context.Context, tx pgx.Tx, list []models.Notification) error {

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"notifications"},
		[]string{
			"id", "group_id", "channel", "recipient",
			"content", "priority",
			"scheduled_at", "status", "created_at",
		},
		pgx.CopyFromSlice(len(list), func(i int) ([]interface{}, error) {
			n := list[i]
			return []interface{}{
				n.Id,
				n.GroupId,
				n.Channel,
				n.Recipient,
				n.Content,
				n.Priority,
				n.ScheduledAt,
				n.Status,
				n.CreatedAt,
			}, nil
		}),
	)

	return err
}

func copyOutbox(ctx context.Context, tx pgx.Tx, list []*models.OutboxEvent) error {

	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"outbox"},
		[]string{
			"id", "aggregate_id",
			"group_id", "event_type", "topic",
			"payload", "status", "retry_count",
			"created_at",
		},
		pgx.CopyFromSlice(len(list), func(i int) ([]interface{}, error) {
			e := list[i]
			return []interface{}{
				e.Id,
				e.AggregateId,
				e.GroupId,
				e.EventType,
				e.Topic,
				e.Payload,
				"pending",
				e.RetryCount,
				e.CreatedAt,
			}, nil
		}),
	)

	return err
}

func (r *PostgresNotificationRepository) FindById(
	ctx context.Context,
	id string,
) (*models.Notification, error) {

	query := `
		SELECT id, recipient, channel, content, status, created_at
		FROM notifications
		WHERE id = $1
	`

	row := r.db.Read.QueryRow(ctx, query, id)

	var n models.Notification
	if err := row.Scan(
		&n.Id,
		&n.Recipient,
		&n.Channel,
		&n.Content,
		&n.Status,
		&n.CreatedAt,
	); err != nil {
		return nil, err
	}

	return &n, nil
}
