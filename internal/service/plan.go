package service

import (
	"context"

	"necutya/faker/internal/domain/domain"
)

type PlanRepository interface {
	GetMany(ctx context.Context) ([]*domain.Plan, error)
	GetOneByName(ctx context.Context, name string) (*domain.Plan, error)
	GetOne(ctx context.Context, id string) (*domain.Plan, error)
}

type PlanService struct {
	planRepo PlanRepository
}

func NewPlanService(
	planRepo PlanRepository,
) *PlanService {
	return &PlanService{
		planRepo: planRepo,
	}
}

func (s *PlanService) GetMany(ctx context.Context) ([]*domain.Plan, error) {
	return s.planRepo.GetMany(ctx)
}

func (s *PlanService) GetOne(ctx context.Context, planID string) (*domain.Plan, error) {
	return s.planRepo.GetOne(ctx, planID)
}

func createPlanMap(plans []*domain.Plan) map[string]*domain.Plan {
	plansMap := make(map[string]*domain.Plan, len(plans))

	for i := range plans {
		plansMap[plans[i].ID] = plans[i]
	}

	return plansMap
}
