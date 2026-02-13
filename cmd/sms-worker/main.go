package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/HuseyinAsik/Notifications/cmd/sms-worker/pkg/settings"
	"github.com/HuseyinAsik/Notifications/pkg/gpostgresql"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/pkg/worker"
	"github.com/HuseyinAsik/Notifications/providers"
	"github.com/HuseyinAsik/Notifications/repository/postgre"
)

func init() {
	settings.Setup()
	logging.Setup(settings.AppSettings)
	logger := logging.GetLogger()
	gpostgresql.Setup(settings.DatabaseSettings, logger)
}

func main() {
	pgPool := gpostgresql.GetPool()
	repo := postgre.NewPostgresNotificationRepository(pgPool)
	logger := logging.GetLogger()
	w := worker.NewWorker(
		settings.KafkaSettings.Brokers,
		"sms",
		100,
		&providers.SMSProvider{},
		repo,
		logger,
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	w.Start(ctx)
}
