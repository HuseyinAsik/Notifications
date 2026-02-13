package routers

import (
	"net/http"

	"github.com/HuseyinAsik/Notifications/cmd/notification-api/pkg/settings"
	"github.com/HuseyinAsik/Notifications/controller"
	"github.com/HuseyinAsik/Notifications/middleware"
	"github.com/HuseyinAsik/Notifications/pkg/gpostgresql"
	"github.com/HuseyinAsik/Notifications/pkg/httpx"
	logging "github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/repository/postgre"
	"github.com/HuseyinAsik/Notifications/services"
	"github.com/gin-gonic/gin"
)

func NewRouter(logger *logging.LogWrapper) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.TimeoutMiddleware(settings.AppSettings.ContextTimeout_))
	r.Use(middleware.LogMiddleware(logger.ZapLogger))
	r.Use(middleware.LogRecoveryMiddleware(logger.ZapLogger))
	r.GET("/healthcheck", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "OK"}) })
	return r
}

func BuildServices(httpClient httpx.HTTPClient, pgPool *gpostgresql.Pool) *gin.Engine {
	logger := logging.GetLogger()
	router := NewRouter(logger)
	repo := postgre.NewPostgresNotificationRepository(pgPool)

	notificationService := services.NewNotificationService(repo, logger)
	controller.NewNotificationController(router, notificationService, logger)

	return router
}
