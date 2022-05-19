package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	core "necutya/faker/internal/domain"
	"necutya/faker/internal/domain/domain"

	"github.com/google/uuid"
)

const (
	verifiedPasswordCheck = "verified"
	codeLength            = 8
	externalTokenLength   = 16
)

type Hasher interface {
	Hash(password string) (string, error)
	CheckPasswordHash(password, hash string) bool
}

type TokenManager interface {
	GenerateAccessToken(userID, planID, sessionID, tokenID string, ttl int64) (string, error)
	GenerateRefreshToken() (string, error)
	Parse(accessToken string) (map[string]interface{}, error)
}

type BlacklistRepository interface {
	AddToken(ctx context.Context, tokenID string, ttl int) error
	CheckToken(ctx context.Context, tokenID string) error
}

type VerificationRepository interface {
	SetCode(context.Context, string, string, int) error
	GetCode(context.Context, string) (string, error)
}

type CodeManager interface {
	GenerateNumericCode(length int) string
	GenerateString(s int) (string, error)
}

type AuthService struct {
	userRepo         UserRepository
	blackListRepo    BlacklistRepository
	reqCounterRepo   RequestCounterRepository
	verificationRepo VerificationRepository
	planRepo         PlanRepository

	hasher              Hasher
	tokenManager        TokenManager
	notificationManager NotificationManager
	codeManager         CodeManager

	accessTokenTTL      int
	refreshTokenTTL     int
	verificationCodeTTL int
}

func NewAuthService(
	userRepo UserRepository,
	blackListRepo BlacklistRepository,
	reqCounterRepo RequestCounterRepository,
	verificationRepo VerificationRepository,
	planRepo PlanRepository,

	hasher Hasher,
	tokenManager TokenManager,
	notificationManager NotificationManager,
	codeManager CodeManager,

	accessTokenTTL int,
	refreshTokenTTL int,
	verificationCodeTTL int,
) *AuthService {
	return &AuthService{
		userRepo:            userRepo,
		blackListRepo:       blackListRepo,
		reqCounterRepo:      reqCounterRepo,
		verificationRepo:    verificationRepo,
		planRepo:            planRepo,
		hasher:              hasher,
		tokenManager:        tokenManager,
		notificationManager: notificationManager,
		codeManager:         codeManager,
		accessTokenTTL:      accessTokenTTL,
		refreshTokenTTL:     refreshTokenTTL,
		verificationCodeTTL: verificationCodeTTL,
	}
}

type UserSignUpInput struct {
	FirstName           string
	LastName            string
	Email               string
	Password            string
	ReceiveNotification bool
}

func (s *AuthService) SignUp(ctx context.Context, input UserSignUpInput) (*domain.User, error) {
	passwordHash, err := s.hasher.Hash(input.Password)
	if err != nil {
		return nil, err
	}

	externalToken, err := s.codeManager.GenerateString(externalTokenLength)
	if err != nil {
		return nil, err
	}

	plan, err := s.planRepo.GetOneByName(ctx, domain.BasicPlanName)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.Create(ctx, &domain.User{
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,

		Role: domain.BasicRole,

		ReceiveNotification: input.ReceiveNotification,
		IsConfirmed:         false,

		Password: passwordHash,

		PlanID: plan.ID,

		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		LastVisitAt: time.Now(),

		Credential: &domain.UserCredential{
			ExternalToken: externalToken,
		},
	})

	verificationCode := s.codeManager.GenerateNumericCode(codeLength)

	err = s.verificationRepo.SetCode(ctx, emailConfirmationKey(input.Email), verificationCode, s.verificationCodeTTL)
	if err != nil {
		return nil, err
	}

	emailConfirmTemplate, err := getEmailConfirmationTemplate(
		strings.Title(fmt.Sprintf("%s %s", input.FirstName, input.LastName)),
		verificationCode,
		int64(s.verificationCodeTTL),
	)
	if err != nil {
		return nil, err
	}

	err = s.notificationManager.SendEmail([]string{input.Email}, domain.ConfirmEmailSubject, emailConfirmTemplate)
	if err != nil {
		return nil, err
	}

	return s.userRepo.GetUserByEmail(ctx, input.Email)
}

type UserSignInInput struct {
	Email     string
	Password  string
	Client    string
	IPAddress string
}

func (s *AuthService) SignIn(ctx context.Context, input UserSignInInput) (*domain.Token, *domain.User, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return nil, nil, core.ErrInvalidLoginOrPassword
		}

		return nil, nil, err
	}

	if !s.hasher.CheckPasswordHash(input.Password, user.Password) {
		return nil, nil, core.ErrInvalidLoginOrPassword
	}

	if !user.IsConfirmed {
		return nil, nil, core.ErrUnconfirmedEmail
	}

	tokens, err := s.createSession(ctx, user.ID, user.PlanID, input.Client, input.IPAddress)
	if err != nil {
		return nil, nil, err
	}

	return tokens, user, nil
}

func (s *AuthService) createSession(ctx context.Context, userID, planID string, client, IPAddress string) (*domain.Token, error) {
	tokenID := uuid.New()
	sessionID := uuid.New()

	accessToken, err := s.tokenManager.GenerateAccessToken(
		userID,
		planID,
		sessionID.String(),
		tokenID.String(),
		int64(s.accessTokenTTL),
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.userRepo.SetSession(ctx, userID, &domain.Session{
		ID:           sessionID,
		RefreshToken: refreshToken,
		Client:       client,
		IpAddress:    IPAddress,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(time.Duration(s.refreshTokenTTL) * time.Second),
	})
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, err
}

func (s *AuthService) SignOut(ctx context.Context, userID string, tokenID, sessionID uuid.UUID) error {
	return s.revokeSession(ctx, userID, tokenID, sessionID)
}

func (s *AuthService) revokeSession(ctx context.Context, userID string, sessionID, tokenID uuid.UUID) error {
	err := s.userRepo.RemoveSession(ctx, userID, sessionID)

	err = s.blackListRepo.AddToken(ctx, tokenID.String(), s.accessTokenTTL)
	if err != nil {
		return err
	}

	return err
}

func (s *AuthService) parseToken(accessToken string) (*domain.TokenInfo, error) {
	claims, err := s.tokenManager.Parse(accessToken)
	if err != nil {
		return nil, err
	}

	return &domain.TokenInfo{
		UserID:    claims["user_id"].(string),
		SessionID: claims["session_id"].(string),
		TokenID:   claims["token_id"].(string),
		PlanID:    claims["plan_id"].(string),
	}, nil
}

func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*domain.TokenInfo, error) {
	tokenInfo, err := s.parseToken(accessToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, tokenInfo.UserID)
	if err != nil {
		return nil, err
	}

	if !user.IsConfirmed {
		return nil, core.ErrUnconfirmedEmail
	}

	if err = s.blackListRepo.CheckToken(ctx, tokenInfo.TokenID); err != nil {
		return nil, err
	}

	return tokenInfo, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, userID, refreshToken string) (*domain.Token, error) {
	session, err := s.userRepo.GetSessionByRefreshToken(ctx, userID, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	err = s.userRepo.RemoveSession(ctx, userID, session.ID)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, core.ErrExpiredSession
	}

	tokenID := uuid.New()

	accessToken, err := s.tokenManager.GenerateAccessToken(
		user.ID,
		user.PlanID,
		session.ID.String(),
		tokenID.String(),
		int64(s.accessTokenTTL),
	)
	if err != nil {
		return nil, err
	}

	return &domain.Token{
		AccessToken:  accessToken,
		RefreshToken: session.RefreshToken,
	}, nil
}

func (s *AuthService) ValidateExternalToken(ctx context.Context, externalToken string) (string, error) {
	user, err := s.userRepo.GetUserByExternalToken(ctx, externalToken)
	if err != nil {
		return "", err
	}

	if !user.IsConfirmed {
		return "", core.ErrUnconfirmedEmail
	}

	return user.ID, nil
}

func (s *AuthService) ConfirmUserEmail(ctx context.Context, userEmail, codeToCheck string) error {
	code, err := s.verificationRepo.GetCode(ctx, emailConfirmationKey(userEmail))
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return core.ErrExpiredCode
		}

		return err
	}

	if code != codeToCheck {
		return core.ErrInvalidCode
	}

	user, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return err
	}

	user.IsConfirmed = true

	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) PasswordResetRequest(ctx context.Context, userEmail string) error {
	user, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return err
	}

	verificationCode := s.codeManager.GenerateNumericCode(codeLength)

	err = s.verificationRepo.SetCode(ctx, passwordResetKey(userEmail), verificationCode, s.verificationCodeTTL)
	if err != nil {
		return err
	}

	passwordResetTemplate, err := getPasswordResetTemplate(
		strings.Title(fmt.Sprintf("%s %s", user.FirstName, user.LastName)),
		verificationCode,
		int64(s.verificationCodeTTL),
	)
	if err != nil {
		return err
	}

	err = s.notificationManager.SendEmail([]string{userEmail}, domain.ConfirmEmailSubject, passwordResetTemplate)
	if err != nil {
		return err
	}

	return nil
}

func (s *AuthService) VerifyPasswordReset(ctx context.Context, userEmail, codeToCheck string) error {
	code, err := s.verificationRepo.GetCode(ctx, passwordResetKey(userEmail))
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return core.ErrExpiredCode
		}

		return err
	}
	log.Println(code)

	if code != codeToCheck {
		return core.ErrInvalidCode
	}

	return s.verificationRepo.SetCode(ctx, passwordResetKey(userEmail), verifiedPasswordCheck, s.verificationCodeTTL)
}

type UserPasswordResetInput struct {
	Password        string
	PasswordConfirm string
}

func (s *AuthService) PasswordReset(ctx context.Context, userEmail string, input UserPasswordResetInput) error {
	code, err := s.verificationRepo.GetCode(ctx, passwordResetKey(userEmail))
	if err != nil {
		if errors.Is(err, core.ErrNotFound) {
			return core.ErrExpiredCode
		}

		return err
	}

	if code != verifiedPasswordCheck {
		return core.ErrUnverifiedPasswordReset
	}

	user, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		return err
	}

	user.Password, err = s.hasher.Hash(input.Password)
	if err != nil {
		return err
	}

	return s.userRepo.Update(ctx, user)
}

func emailConfirmationKey(email string) string {
	return fmt.Sprintf("confirm:%s", email)
}

func passwordResetKey(email string) string {
	return fmt.Sprintf("password-reset:%s", email)
}
