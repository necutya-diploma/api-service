package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	core "necutya/faker/internal/domain"
	"necutya/faker/internal/domain/domain"
	"necutya/faker/pkg/logger"
	"necutya/faker/pkg/payments/fondy"

	"github.com/google/uuid"
)

type PaymentsManager interface {
	GenerateSubscriptionCheckout(
		orderID, orderDesc, senderEmail, productID string,
		amount int64,
	) (string, error)
	ValidateCallback(input interface{}) error
}

type PaymentsService struct {
	orderRepo OrderRepository
	userRepo  UserRepository
	planRepo  PlanRepository

	paymentsManager     PaymentsManager
	notificationManager NotificationManager
}

func NewPaymentsService(
	orderRepo OrderRepository,
	userRepo UserRepository,
	planRepo PlanRepository,
	paymentsManager PaymentsManager,
	notificationManager NotificationManager,
) *PaymentsService {
	return &PaymentsService{
		paymentsManager:     paymentsManager,
		orderRepo:           orderRepo,
		userRepo:            userRepo,
		planRepo:            planRepo,
		notificationManager: notificationManager,
	}
}

func (s *PaymentsService) GetPaymentsCheckout(ctx context.Context) (string, error) {
	return s.paymentsManager.GenerateSubscriptionCheckout(
		uuid.New().String(),
		"test",
		"art.lebedev2020@gmail.com",
		uuid.New().String(),
		1000,
	)
}

func (s *PaymentsService) ProcessCallback(ctx context.Context, callback interface{}) error {
	switch callbackData := callback.(type) {
	case fondy.Callback:
		return s.processFondyCallback(ctx, callbackData)
	default:
		return core.ErrUnknownCallbackType
	}
}

func (s *PaymentsService) processFondyCallback(ctx context.Context, callback fondy.Callback) error {
	orderID := callback.OrderId

	log.Println(callback)

	if err := s.paymentsManager.ValidateCallback(callback); err != nil {
		return core.ErrTransactionInvalid
	}

	transaction, err := createTransactionFromFondyCallback(&callback)
	if err != nil {
		return err
	}

	_, err = s.orderRepo.AddTransaction(ctx, orderID, transaction)
	if err != nil {
		return err
	}

	if transaction.Status != domain.PaidOrderStatus {
		return nil
	}

	user, err := s.userRepo.GetUserByEmail(ctx, callback.SenderEmail)
	if err != nil {
		return err
	}

	user.PlanID = callback.ProductId

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	plan, err := s.planRepo.GetOne(ctx, callback.ProductId)
	if err != nil {
		log.Println("get plan")
		return err
	}

	if err = s.sendSuccessNotification(user, plan); err != nil {
		logger.Errorf("failed to send email after purchase: %s", err.Error())
	}

	return nil
}

func (s *PaymentsService) sendSuccessNotification(user *domain.User, plan *domain.Plan) error {
	orderSuccessTemplate, err := getOrderSuccessTemplate(
		strings.Title(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		plan.Name,
		plan.EndDateFromNow(),
	)
	if err != nil {
		return err
	}

	return s.notificationManager.SendEmail([]string{user.Email}, domain.SuccessEmailSubject, orderSuccessTemplate)
}

func createTransactionFromFondyCallback(callback *fondy.Callback) (*domain.Transaction, error) {
	var status string
	if callback.PaymentApproved() {
		status = domain.PaidOrderStatus
	} else {
		status = domain.OtherOrderStatus
	}

	if !callback.Success() {
		status = domain.FailedOrderStatus
	}

	additionalInfo, err := json.Marshal(callback)
	if err != nil {
		return nil, err
	}

	return &domain.Transaction{
		Status:         status,
		CreatedAt:      time.Now(),
		AdditionalInfo: string(additionalInfo),
	}, nil
}
