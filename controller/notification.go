package controller

import (
	"net/http"
	"time"

	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/HuseyinAsik/Notifications/serializers"
	"github.com/HuseyinAsik/Notifications/services"
	"github.com/gin-gonic/gin"
)

type notificationController struct {
	Logger              *logging.LogWrapper
	NotificationService *services.NotificationService
}

func NewNotificationController(R *gin.Engine, notificationService *services.NotificationService, logger *logging.LogWrapper) {

	controller := &notificationController{
		NotificationService: notificationService,
		Logger:              logger,
	}

	api := R.Group("api/v1/notifications")
	{
		api.POST("", controller.Create)
		api.POST("/batch", controller.Batch)
		api.GET("", controller.List)
	}
}

func (c *notificationController) Create(g *gin.Context) {
	serializer := serializers.Serializer{C: g, Logger: c.Logger}
	ctx := g.Request.Context()
	var form serializers.CreateNotificationForm

	_ = serializer.ShouldBindJSON(ctx, &form)
	if validateErr := form.Validate(ctx); validateErr != nil {
		serializer.ErrorResponse(http.StatusBadRequest, validateErr)
		return
	}
	Id, createErr := c.NotificationService.Create(ctx, &form)

	if createErr != nil {
		serializer.ErrorResponse(http.StatusInternalServerError, createErr)
	}

	serializer.NotificationResponse(http.StatusAccepted, serializers.NotificationResponse{
		MessageId: Id,
		Status:    "Accepted",
		CreatedAt: time.Now().Format(time.RFC3339),
	})
}

func (c *notificationController) Batch(g *gin.Context) {
	serializer := serializers.Serializer{C: g, Logger: c.Logger}
	ctx := g.Request.Context()
	var form serializers.CreateNotificationBatchForm

	_ = serializer.ShouldBindJSON(ctx, &form)
	if validateErr := form.Validate(ctx); validateErr != nil {
		serializer.ErrorResponse(http.StatusBadRequest, validateErr)
		return
	}

	Id, bulkErr := c.NotificationService.BulkCreate(ctx, form)

	if bulkErr != nil {
		serializer.ErrorResponse(http.StatusInternalServerError, bulkErr)
	}

	serializer.NotificationResponse(http.StatusAccepted, serializers.NotificationResponse{
		MessageId: Id,
		Status:    "Accepted",
		CreatedAt: time.Now().Format(time.RFC3339),
	})
}

func (c *notificationController) List(g *gin.Context) {
	serializer := serializers.Serializer{C: g, Logger: c.Logger}
	ctx := g.Request.Context()
	var form serializers.ListForm
	_ = serializer.ShouldBindQuery(ctx, &form)

	if err := form.Validate(ctx); err != nil {
		serializer.ErrorResponse(http.StatusBadRequest, err)
		return
	}

	notifications, total, err := c.NotificationService.List(ctx, form)

	if err != nil {
		serializer.ErrorResponse(http.StatusInternalServerError, err)
	}

	serializer.NotificationListResponse(http.StatusOK, serializers.NotificationListResponse{
		Notifications: notifications,
		Total:         total,
	})
}
