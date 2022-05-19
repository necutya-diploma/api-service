package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	core "necutya/faker/internal/domain"
	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/domain/dto"
	"necutya/faker/pkg/logger"

	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) (*domain.Order, error)
	GetByID(ctx context.Context, userID string) (*domain.Order, error)
	AddTransaction(
		ctx context.Context,
		id string,
		transaction *domain.Transaction,
	) (*domain.Order, error)
	GetByUserID(ctx context.Context, userID string) ([]*domain.Order, error)
}

type NotificationManager interface {
	SendEmail(to []string, subject, body string) error
}

type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetUserByExternalToken(ctx context.Context, externalToken string) (*domain.User, error)

	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID string) error

	SetSession(ctx context.Context, userID string, session *domain.Session) error
	GetSessionByRefreshToken(ctx context.Context, userID string, refreshToken string) (*domain.Session, error)
	RemoveSession(ctx context.Context, userID string, sessionID uuid.UUID) error

	GetMany(context.Context, dto.UserFilter) ([]*domain.User, error)
}

type UserService struct {
	userRepo            UserRepository
	planRepo            PlanRepository
	orderRepo           OrderRepository
	requestCounterRepo  RequestCounterRepository
	hasher              Hasher
	notificationManager NotificationManager
	paymentsManager     PaymentsManager
	codeManager         CodeManager
}

func NewUserService(
	userRepo UserRepository,
	planRepo PlanRepository,
	orderRepo OrderRepository,
	requestCounterRepo RequestCounterRepository,
	hasher Hasher,
	notificationManager NotificationManager,
	paymentsManager PaymentsManager,
	codeManager CodeManager,
) *UserService {
	return &UserService{
		userRepo:            userRepo,
		planRepo:            planRepo,
		orderRepo:           orderRepo,
		requestCounterRepo:  requestCounterRepo,
		hasher:              hasher,
		notificationManager: notificationManager,
		paymentsManager:     paymentsManager,
		codeManager:         codeManager,
	}
}

func (s *UserService) GetOne(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.TodayExternalRequest, err = s.requestCounterRepo.GetByUserID(ctx, userID, domain.External)
	if err != nil {
		return nil, err
	}

	user.TodayInternalRequest, err = s.requestCounterRepo.GetByUserID(ctx, userID, domain.Internal)
	if err != nil {
		return nil, err
	}

	return user, err
}

type UserUpdateInput struct {
	FirstName string
	LastName  string
}

func (s *UserService) Update(ctx context.Context, userID string, input UserUpdateInput) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.FirstName = input.FirstName
	user.LastName = input.LastName

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.GetOne(ctx, user.ID)
}

func (s *UserService) Delete(ctx context.Context, userID string) error {
	return s.userRepo.Delete(ctx, userID)
}

func (s *UserService) UpdateNotification(ctx context.Context, userID string, receive bool) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.ReceiveNotification = receive

	err = s.userRepo.Update(ctx, user)

	return s.GetOne(ctx, user.ID)
}

type UserUpdatePasswordInput struct {
	CurrentPassword string
	NewPassword     string
}

func (s *UserService) UpdateUserPassword(ctx context.Context, userID string, input UserUpdatePasswordInput) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	if !s.hasher.CheckPasswordHash(input.CurrentPassword, user.Password) {
		return core.ErrInvalidCurrentPassword
	}

	user.Password, err = s.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) UpdateUserPlan(ctx context.Context, userID, planID string) (string, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if user.PlanID == planID {
		return "", core.ErrThisPlanAlreadySet
	}

	plan, err := s.planRepo.GetOne(ctx, planID)
	if err != nil {
		return "", err
	}

	if plan.IsBasic() {
		user.PlanID = plan.ID
		return "", s.userRepo.Update(ctx, user)
	}

	now := time.Now()

	orderDesc := domain.GenerateOrderDescription(
		strings.Title(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		fmt.Sprintf("Plan \"%s\"", plan.Name),
		plan.Price,
		now,
	)

	order := &domain.Order{
		UserID:      userID,
		PlanID:      planID,
		Amount:      plan.Price,
		CreatedAt:   now,
		Description: orderDesc,
		Status:      domain.NewOrderStatus,
	}

	order, err = s.orderRepo.Create(ctx, order)
	if err != nil {
		return "", err
	}

	checkoutLink, err := s.paymentsManager.GenerateSubscriptionCheckout(
		order.ID,
		order.Description,
		user.Email,
		plan.ID,
		int64(convertDollarsToCents(plan.Price)),
	)
	if err != nil {
		return "", err
	}

	return checkoutLink, nil
}

func (s *UserService) RegenerateExternalToken(ctx context.Context, userID string) (*domain.User, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	user.Credential.ExternalToken, err = s.codeManager.GenerateString(externalTokenLength)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, err
	}

	return s.GetOne(ctx, user.ID)
}

func (s *UserService) ValidateUsersPlan(ctx context.Context) error {
	var lastPaidOrder *domain.Order

	basicPlan, err := s.planRepo.GetOneByName(ctx, domain.BasicPlanName)
	if err != nil {
		return err
	}

	users, err := s.userRepo.GetMany(ctx, dto.UserFilter{
		Role:        string(domain.BasicRole),
		IsConfirmed: true,
	})
	if err != nil {
		return err
	}

	for i := range users {
		lastPaidOrder = nil

		if users[i].PlanID == basicPlan.ID {
			continue
		}

		orders, err := s.orderRepo.GetByUserID(ctx, users[i].ID)
		if err != nil {
			return err
		}

		sort.Slice(orders, func(i, j int) bool {
			return orders[i].CreatedAt.After(orders[j].CreatedAt)
		})

		for j := range orders {
			if orders[j].Paid() {
				lastPaidOrder = orders[j]
				break
			}
		}

		if lastPaidOrder != nil && time.Now().Sub(lastPaidOrder.CreatedAt).Hours() > float64(domain.HoursInMonth()) {
			users[i].PlanID = basicPlan.ID
			if err = s.sendUpdatePlan(users[i]); err != nil {
				return err
			}
		}

		err = s.userRepo.Update(ctx, users[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *UserService) ValidateUsersPlanSync() {
	ctx := context.Background()
	traceID := "ValidatePlanSync"
	logger.Infof(fmt.Sprintf("%s starts", traceID))

	if err := s.ValidateUsersPlan(ctx); err != nil {
		err = errors.Wrap(errors.WithStack(err), fmt.Sprintf("TraceID: %+v", traceID))
		logger.Error(err)
	}

	logger.Infof(fmt.Sprintf("%s ends", traceID))
}

func (s *UserService) sendUpdatePlan(user *domain.User) error {
	planUpdateTmpl, err := getPlanUpdateTemplate(
		strings.Title(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
	)
	if err != nil {
		return err
	}

	return s.notificationManager.SendEmail([]string{user.Email}, domain.DeactivatedPlanSubject, planUpdateTmpl)
}
