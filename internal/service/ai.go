package service

import (
	"context"
	"time"

	core "necutya/faker/internal/domain"
	"necutya/faker/internal/domain/domain"
)

type MessageRepository interface {
	Create(context.Context, *domain.Message) error
}

type RequestCounterRepository interface {
	Incr(context.Context, string, domain.RequestType) error
	GetByUserID(context.Context, string, domain.RequestType) (int, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
}

type AIManager interface {
	CheckMessage(msg string) (string, bool, float64, error)
}

type AIService struct {
	messageRepo        MessageRepository
	requestCounterRepo RequestCounterRepository
	aiManager          AIManager
	planRepo           PlanRepository
}

func NewAIService(
	messageRepo MessageRepository,
	requestCounterRepo RequestCounterRepository,
	aiManager AIManager,
	planRepo PlanRepository,
) *AIService {
	return &AIService{
		messageRepo:        messageRepo,
		requestCounterRepo: requestCounterRepo,
		aiManager:          aiManager,
		planRepo:           planRepo,
	}
}

type CheckMessageInput struct {
	Message  string
	SaveToDb bool
}

type CheckMessageOutput struct {
	Message          string
	IsGenerated      bool
	GeneratedPercent float64
}

func (s *AIService) CheckMessage(ctx context.Context, userID, planID string, input *CheckMessageInput, requestType domain.RequestType) (*CheckMessageOutput, error) {
	var (
		err    error
		output = CheckMessageOutput{}
	)

	if err = s.validateRequestCountPerDay(ctx, userID, planID, requestType); err != nil {
		return nil, err
	}

	output.Message, output.IsGenerated, output.GeneratedPercent, err = s.aiManager.CheckMessage(input.Message)
	if err != nil {
		return nil, err
	}

	if input.SaveToDb {
		err = s.messageRepo.Create(ctx, &domain.Message{
			Text:      input.Message,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
		if err != nil {
			return nil, err
		}
	}

	err = s.incrementRequestCounter(ctx, userID, requestType)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (s *AIService) incrementRequestCounter(ctx context.Context, userID string, requestType domain.RequestType) error {
	reqCount, err := s.requestCounterRepo.GetByUserID(ctx, userID, requestType)
	if err != nil {
		return err
	}

	err = s.requestCounterRepo.Incr(ctx, userID, requestType)
	if err != nil {
		return err
	}

	if reqCount == 0 {
		return s.requestCounterRepo.Expire(ctx, userID, getDurationToMidnight())
	}

	return nil
}

func (s *AIService) validateRequestCountPerDay(ctx context.Context, userID, planID string, requestType domain.RequestType) error {
	reqCount, err := s.requestCounterRepo.GetByUserID(ctx, userID, requestType)
	if err != nil {
		return err
	}

	plan, err := s.planRepo.GetOne(ctx, planID)
	if err != nil {
		return err
	}

	switch requestType {
	case domain.External:
		if plan.ExternalRequestsCount != 0 && reqCount > plan.ExternalRequestsCount {
			return core.ErrRequestLimit
		}
	case domain.Internal:
		if plan.InternalRequestsCount != 0 && reqCount > plan.InternalRequestsCount {
			return core.ErrRequestLimit
		}
	}

	return nil
}
