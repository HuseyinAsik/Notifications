package models

import "time"

type Notification struct {
	Id          string     `json:"id,omitempty"`
	GroupId     string     `json:"groupId,omitempty"`
	Recipient   string     `json:"recipient,omitempty"`
	Channel     string     `json:"channel,omitempty"`
	Content     string     `json:"content,omitempty"`
	Status      string     `json:"status,omitempty"`
	Priority    string     `json:"priority,omitempty"`
	ScheduledAt *time.Time `json:"scheduledAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt,omitempty"`
}
