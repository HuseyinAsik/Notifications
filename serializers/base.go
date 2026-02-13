package serializers

import (
	"context"

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

func (s *Serializer) ShouldBindJSON(ctx context.Context, obj interface{}) error {
	err := s.C.ShouldBindJSON(obj)
	if err != nil {
		s.Logger.Warn(ctx, "Serializer ShouldBindJSON Validation Err", zap.Error(err))
		return err
	}
	return nil
}

func (s *Serializer) Validate(ctx context.Context, form interface{}) error {
	validate := validator.New()
	err := validate.StructCtx(ctx, form)

	return err
}
func (s *Serializer) ResponseNoData(httpCode int) {
	s.C.Status(httpCode)
}
func (s *Serializer) ErrorResponse(httpCode int, err error) {
	s.C.JSON(httpCode, gin.H{"errorDetail": err.Error()})
}
func (s *Serializer) NotificationResponse(httpCode int, data NotificationResponse) {
	s.C.JSON(httpCode, data)
}
