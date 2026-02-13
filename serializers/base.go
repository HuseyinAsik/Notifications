package serializers

import (
	"context"

	"github.com/HuseyinAsik/Notifications/models"
	"github.com/HuseyinAsik/Notifications/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type Serializer struct {
	C      *gin.Context
	Logger *logging.LogWrapper
}

type NotificationResponse struct {
	MessageId string `json:"messageId"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type NotificationListResponse struct {
	Total         int                   `json:"messageId"`
	Notifications []models.Notification `json:"notifications"`
}

func (s *Serializer) ShouldBindJSON(ctx context.Context, obj interface{}) error {
	err := s.C.ShouldBindJSON(obj)
	if err != nil {
		s.Logger.Warn(ctx, "Serializer ShouldBindJSON Validation Err", zap.Error(err))
		return err
	}
	return nil
}

func (s *Serializer) ShouldBindQuery(ctx context.Context, obj interface{}) error {
	err := s.C.ShouldBindQuery(obj)
	if err != nil {
		s.Logger.Warn(ctx, "Serializer ShouldBindQuery Validation Err", zap.Error(err))
		return err
	}
	return nil
}

func (s *Serializer) Validate(ctx context.Context, form interface{}) error {
	validate := validator.New()
	err := validate.StructCtx(ctx, form)

	return err
}
func (s *Serializer) ErrorResponse(httpCode int, err error) {
	s.C.JSON(httpCode, gin.H{"errorDetail": err.Error()})
}
func (s *Serializer) NotificationResponse(httpCode int, data NotificationResponse) {
	s.C.JSON(httpCode, data)
}

func (s *Serializer) NotificationListResponse(httpCode int, data NotificationListResponse) {
	s.C.JSON(httpCode, data)
}
