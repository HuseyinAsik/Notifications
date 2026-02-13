package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HuseyinAsik/Notifications/cmd/notification-api/pkg/settings"
	"github.com/HuseyinAsik/Notifications/pkg/gpostgresql"
	"github.com/HuseyinAsik/Notifications/pkg/httpx"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/routers"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func init() {
	settings.Setup()
	logging.Setup(settings.AppSettings)
	logger := logging.GetLogger()
	gpostgresql.Setup(settings.DatabaseSettings, logger)
}

func main() {
	logger := logging.GetLogger()
	gin.SetMode(settings.AppSettings.RunMode)
	if settings.AppSettings.RunMode == gin.ReleaseMode {
		gin.DefaultWriter = io.Discard
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	httpxClient := httpx.NewHTTPClient(client, logger)
	postgresqlPool := gpostgresql.GetPool()
	router := routers.BuildServices(httpxClient, postgresqlPool)

	readTimeout := time.Second
	writeTimeout := time.Second
	endPoint := fmt.Sprintf(":%s", settings.ServerSettings.HttpPort)
	server := &http.Server{
		Addr:           endPoint,
		Handler:        router,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}
	ctx := context.Background()

	logger.Info(ctx, "Start http server listening", zap.String("endPoint", endPoint))
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error(ctx, "Server listening error", zap.Error(err))
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	<-quit
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	defer func() {
		_ = logger.Sync()
	}()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error(ctx, "Server shutdown error", zap.Error(err))
	}
	// db, err := sql.Open(
	// 	"postgres",
	// 	"postgres://user:password@localhost:5432/notifications?sslmode=disable",
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// repo := postgre.NewPostgresNotificationRepository(db)
	// svc := services.NewNotificationService(repo)

	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	// n := &models.Notification{
	// 	Recipient: "+905551234567",
	// 	Channel:   "sms",
	// 	Content:   "Hello Repository Pattern!",
	// }

	// if err := svc.CreateNotification(ctx, n); err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("Created notification:", n.ID)
}
