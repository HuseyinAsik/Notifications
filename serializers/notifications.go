package serializers

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type CreateNotificationForm struct {
	Recipient   string     `json:"recipient" validate:"required"`
	Channel     string     `json:"channel" validate:"required,oneof=sms email push"`
	Content     string     `json:"content" validate:"required"`
	Priority    string     `json:"priority" validate:"required,oneof=high medium low"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
}

func (s *CreateNotificationForm) Validate(ctx context.Context) error {
	validate := validator.New()
	s.Channel = strings.ToLower(s.Channel)
	s.Priority = strings.ToLower(s.Priority)
	err := validate.StructCtx(ctx, s)

	return err
}

type CreateNotificationBatchForm struct {
	Data []CreateNotificationForm `json:"data" validate:"required,min=1,max=1000,dive"`
}

func (s *CreateNotificationBatchForm) Validate(ctx context.Context) error {
	validate := validator.New()
	err := validate.StructCtx(ctx, s)

	if err != nil {
		return err
	}

	if len(s.Data) == 0 || len(s.Data) > 1000 {
		return errors.New("request must include between 1 to 1000 notifications")
	}

	for i := range s.Data {
		s.Data[i].Channel = strings.ToLower(s.Data[i].Channel)
		s.Data[i].Priority = strings.ToLower(s.Data[i].Priority)
	}

	return nil
}

type ListForm struct {
	PageStr      string `form:"page"`
	Status       string `form:"status"`
	Channel      string `form:"channel"`
	StartDateStr string `form:"startdate"`
	EndDateStr   string `form:"enddate"`
	Page         int
	StartDate    *time.Time
	EndDate      *time.Time
}

func (s *ListForm) Validate(ctx context.Context) error {
	page, err := strconv.Atoi(s.PageStr)
	s.Page = 1
	if err != nil && page > 0 {
		s.Page = page
	}

	if s.StartDateStr != "" {
		if t, err := time.Parse(time.RFC3339, s.StartDateStr); err == nil {
			s.StartDate = &t
		}
	}
	if s.EndDateStr != "" {
		if t, err := time.Parse(time.RFC3339, s.EndDateStr); err == nil {
			s.EndDate = &t
		}
	}

	return nil
}
