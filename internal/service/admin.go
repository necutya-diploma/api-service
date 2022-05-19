package service

import (
	"context"

	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/domain/dto"
)

type UserAdminRepository interface {
	GetMany(context.Context, dto.UserFilter) ([]*domain.User, error)
}

type AdminService struct {
	userRepo UserAdminRepository
	planRepo PlanRepository
}

func NewAdminService(
	userRepo UserAdminRepository,
	planRepo PlanRepository,
) *AdminService {
	return &AdminService{
		userRepo: userRepo,
		planRepo: planRepo,
	}
}

func (s *AdminService) GetUsersReport(ctx context.Context) (*domain.UsersReport, error) {
	users, err := s.userRepo.GetMany(ctx, dto.UserFilter{
		Role:        string(domain.BasicRole),
		IsConfirmed: true,
	})
	if err != nil {
		return nil, err
	}

	totalInfo := domain.TotalInfoForReport{}
	usersInfo := make([]domain.UserInfoForReport, len(users))

	plans, err := s.planRepo.GetMany(ctx)
	if err != nil {
		return nil, err
	}
	plansMap := createPlanMap(plans)

	for i := range users {
		// TODO: add calculations
		usersInfo[i] = domain.UserInfoForReport{
			ID:        users[i].ID,
			FirstName: users[i].FirstName,
			LastName:  users[i].LastName,
			Email:     users[i].Email,
			Plan:      plansMap[users[i].PlanID].Name,
			GPM:       0.,
			Feedbacks: []string{},
		}

		totalInfo.UsersAmount++
		totalInfo.GPM += 0
	}

	return &domain.UsersReport{
		TotalInfo: totalInfo,
		UsersInfo: usersInfo,
	}, nil
}
