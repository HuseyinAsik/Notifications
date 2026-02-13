package models

import (
	"time"
)

type OutboxEvent struct {
	Id          string
	AggregateId string
	GroupId     string
	EventType   string
	Topic       string
	Payload     []byte

	Status      string
	RetryCount  int
	CreatedAt   time.Time
	PublishedAt *time.Time
}
