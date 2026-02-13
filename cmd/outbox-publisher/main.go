package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HuseyinAsik/Notifications/cmd/outbox-publisher/pkg/settings"
	"github.com/HuseyinAsik/Notifications/pkg/gpostgresql"
	"github.com/HuseyinAsik/Notifications/pkg/kafka"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/services"

	"github.com/HuseyinAsik/Notifications/repository/postgre"
)

func init() {
	settings.Setup()
	logging.Setup(settings.AppSettings)
	logger := logging.GetLogger()
	gpostgresql.Setup(settings.DatabaseSettings, logger)
}

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger := logging.GetLogger()
	postgresqlPool := gpostgresql.GetPool()

	repo := postgre.NewPostgresNotificationRepository(postgresqlPool)
	writer := kafka.NewWriter(settings.KafkaSettings.Brokers)

	pub := services.NewOutbox(repo, writer, logger)

	pub.Run(ctx)
}
