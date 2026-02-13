package settings

import (
	"strconv"
	"time"
)

type App struct {
	AppName         string `notification_api_validate:"required" email_worker_validate:"required" sms_worker_validate:"required" push_worker_validate:"required" outbox_publisher_validate:"required"`
	ContextTimeout  string
	ContextTimeout_ time.Duration
	Encoding        string `notification_api_validate:"required" email_worker_validate:"required" sms_worker_validate:"required" push_worker_validate:"required" outbox_publisher_validate:"required"`
	LogLevel        string `notification_api_validate:"required" email_worker_validate:"required" sms_worker_validate:"required" push_worker_validate:"required" outbox_publisher_validate:"required"`
	RunMode         string `notification_api_validate:"required"`
	Hostname        string `notification_api_validate:"required" email_worker_validate:"required" sms_worker_validate:"required" push_worker_validate:"required" outbox_publisher_validate:"required"`
}

func (s *App) Load() {
	ContextTimeout, _ := strconv.Atoi(s.ContextTimeout)
	s.ContextTimeout_ = time.Duration(ContextTimeout)
}

type Server struct {
	HttpPort string `notification_api_validate:"required"`
}

type Database struct {
	ReadUrl  string `notification_api_validate:"required" outbox_publisher_validate:"required"`
	WriteUrl string `notification_api_validate:"required" outbox_publisher_validate:"required"`
}

type Kafka struct {
	BrokersStr string `email_worker_validate:"required" sms_worker_validate:"required" push_worker_validate:"required" outbox_publisher_validate:"required"`
	Brokers    []string
}
