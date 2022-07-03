package service

type Service struct {
	User     *UserService
	AI       *AIService
	Auth     *AuthService
	Admin    *AdminService
	Payments *PaymentsService
	Plan     *PlanService
	Feedback *FeedbackService
}

func New(
	userRepo UserRepository,
	messageRepo MessageRepository,
	planRepo PlanRepository,
	orderRepo OrderRepository,
	feedbackRepo FeedbackRepository,
	blackListRepo BlacklistRepository,
	requestCounterRepo RequestCounterRepository,
	verificationRepo VerificationRepository,

	hasher Hasher,
	tokenManager TokenManager,
	aiManager AIManager,
	notificationManager NotificationManager,
	codeManager CodeManager,
	paymentsManager PaymentsManager,

	accessTokenTTL int,
	refreshTokenTTL int,
	verificationCodeTTL int,

	feedbackReceiver string,
) *Service {
	return &Service{
		Auth: NewAuthService(
			userRepo, blackListRepo, requestCounterRepo, verificationRepo, planRepo,
			hasher, tokenManager, notificationManager, codeManager,
			accessTokenTTL, refreshTokenTTL, verificationCodeTTL,
		),
		User:     NewUserService(userRepo, planRepo, orderRepo, requestCounterRepo, hasher, notificationManager, paymentsManager, codeManager),
		AI:       NewAIService(messageRepo, requestCounterRepo, aiManager, planRepo, userRepo),
		Admin:    NewAdminService(userRepo, planRepo),
		Payments: NewPaymentsService(orderRepo, userRepo, planRepo, paymentsManager, notificationManager),
		Plan:     NewPlanService(planRepo),
		Feedback: NewFeedbackService(feedbackRepo, notificationManager, feedbackReceiver),
	}
}
