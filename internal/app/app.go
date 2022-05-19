package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"necutya/faker/internal/config"
	"necutya/faker/internal/repositories/mongo"
	"necutya/faker/internal/service"
	aiGrpcClient "necutya/faker/pkg/ai-grpc-client"
	"necutya/faker/pkg/generators"
	notificationGrpcClient "necutya/faker/pkg/notification-grpc-client"
	"necutya/faker/pkg/payments/fondy"

	"necutya/faker/internal/repositories/redis"
	"necutya/faker/pkg/database/mongodb"
	redisdb "necutya/faker/pkg/database/redis"
	"necutya/faker/pkg/hasher"
	"necutya/faker/pkg/logger"

	httpTransport "necutya/faker/internal/transport/http"
	httpHandler "necutya/faker/internal/transport/http/handlers"
	tokenManager "necutya/faker/pkg/token_managers"

	"github.com/go-playground/validator/v10"
	"github.com/robfig/cron/v3"
)

func Run(configPath string) {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go gracefulShutdown(cancel)

	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Fatal("can`t connect to mongo:", err.Error())
	}

	redisClient, err := redisdb.NewClient(cfg.Redis.URI, cfg.Redis.Password, cfg.Redis.PoolSize)
	if err != nil {
		logger.Fatal("can`t connect to redis:", err.Error())
	}

	database := mongoClient.Database(cfg.Mongo.DatabaseName)

	jwtTokenManager, err := tokenManager.NewJWT(cfg.Token.SignKey)

	paymentsManager := fondy.NewPaymentManager(
		int64(cfg.Payments.MerchantID),
		cfg.Payments.MerchantPassword,
		cfg.Payments.Currency,
		cfg.Payments.Language,
		cfg.Payments.ResponseURL,
		cfg.Payments.CallbackURL,
	)

	services := service.New(
		mongo.NewUsersRepo(database),
		mongo.NewMessagesRepo(database),
		mongo.NewPlansRepo(database),
		mongo.NewOrdersRepo(database),
		mongo.NewUsersRepo(database),
		redis.NewBlacklistRepo(redisClient),
		redis.NewRequestCounterRepo(redisClient),
		redis.NewVerificationRepo(redisClient),
		hasher.NewBcryptHasher(),
		jwtTokenManager,
		aiGrpcClient.New(cfg.AI.Addr),
		notificationGrpcClient.New(cfg.Notification.Addr, cfg.Notification.From),
		generators.NewRandomGenerator(),
		paymentsManager,
		cfg.Token.AccessTokenTTL,
		cfg.Token.RefreshTokenTTL,
		cfg.VerificationCodeTTL,
		cfg.Feedbacks.Receiver,
	)

	initCronJobs(cfg.Cron, services)

	handlers := httpHandler.New(services, initValidator())

	httpServer := httpTransport.NewHttp(
		&cfg.HTTP,
		handlers.Init(
			cfg.HTTP.URLPrefix,
			cfg.HTTP.ExternalURLPrefix,
			cfg.HTTP.CORSAllowedHost,
		),
	)
	httpServer.Run(ctx)
}

func gracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	<-signalChannel
	stop()
}

func initValidator() *validator.Validate {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return validate
}

func initCronJobs(cfg config.CronConfigs, services *service.Service) {
	c := cron.New()

	if _, err := c.AddFunc(cfg.ValidatePlanSync, services.User.ValidateUsersPlanSync); err != nil {
		logger.Fatal(err, "cron ValidatePlanSync init failed")
	}

	c.Start()
}
