package models

import "time"

type Notification struct {
	Id          string
	GroupId     string
	Recipient   string
	Channel     string
	Content     string
	Status      string
	Priority    string
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	CreatedAt   time.Time
}
