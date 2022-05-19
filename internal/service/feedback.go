package service

import (
	"context"
	"time"

	"necutya/faker/internal/domain/domain"

	"github.com/google/uuid"
)

type FeedbackRepository interface {
	GetUserFeedbacks(ctx context.Context, userID string) ([]*domain.Feedback, error)
	SetUserFeedback(ctx context.Context, userID string, feedback *domain.Feedback) error
	ResolveFeedback(ctx context.Context, userID string, feedbackID uuid.UUID) error
}

type FeedbackService struct {
	feedbackRepo        FeedbackRepository
	notificationManager NotificationManager
	receiverEmail       string
}

func NewFeedbackService(
	feedbackRepo FeedbackRepository,
	notificationManager NotificationManager,
	receiverEmail string,
) *FeedbackService {
	return &FeedbackService{
		feedbackRepo:        feedbackRepo,
		notificationManager: notificationManager,
		receiverEmail:       receiverEmail,
	}
}

func (s *FeedbackService) GetUserFeedbacks(ctx context.Context, userID string) ([]*domain.Feedback, error) {
	return s.feedbackRepo.GetUserFeedbacks(ctx, userID)
}

type FeedbackCreateInput struct {
	UserID string

	Username string
	Email    string
	Text     string
}

func (s *FeedbackService) SetFeedback(ctx context.Context, input FeedbackCreateInput) error {
	feedback := domain.Feedback{
		ID:        uuid.New(),
		Text:      input.Text,
		Processed: false,
		CreatedAt: time.Now(),
		Email:     input.Email,
		Username:  input.Username,
		UserID:    input.UserID,
	}

	if feedback.UserID != "" {
		err := s.feedbackRepo.SetUserFeedback(ctx, feedback.UserID, &feedback)
		if err != nil {
			return err
		}
	}

	return s.notificationManager.SendEmail(
		[]string{s.receiverEmail},
		feedback.GenerateEmailSubject(),
		feedback.GenerateEmailBody(),
	)
}

func (s *FeedbackService) ResolveFeedback(ctx context.Context, userID string, feedbackID string) error {
	feedbackUUID, err := uuid.Parse(feedbackID)
	if err != nil {
		return err
	}

	return s.feedbackRepo.ResolveFeedback(ctx, userID, feedbackUUID)
}
