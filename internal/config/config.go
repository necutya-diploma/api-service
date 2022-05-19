package config

import "time"

type Config struct {
	VerificationCodeTTL int                       `json:"verification_code_ttl"`
	ENV                 string                    `json:"env"`
	Log                 Logger                    `json:"log"`
	HTTP                HTTPConfig                `json:"http"`
	Redis               RedisConfig               `json:"redis"`
	Mongo               MongoConfig               `json:"mongo"`
	Token               TokenConfig               `json:"token"`
	AI                  AiServiceConfig           `json:"ai"`
	Notification        NotificationServiceConfig `json:"notification"`
	Payments            PaymentsConfig            `json:"payments"`
	Feedbacks           FeedbacksConfig           `json:"feedbacks"`
	Cron                CronConfigs               `json:"cron"`
}

type Logger struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type HTTPConfig struct {
	Address            string        `json:"address"`
	URLPrefix          string        `json:"url_prefix"`
	ExternalURLPrefix  string        `json:"external_url_prefix"`
	CSRFSecuredCookie  bool          `json:"csrf_secured_cookie"`
	ReadTimeout        time.Duration `json:"read_timeout"`
	WriteTimeout       time.Duration `json:"write_timeout"`
	MaxHeaderMegabytes int           `json:"max_header_bytes"`
	CORSAllowedHost    []string      `json:"cors_allowed_host"`
}

type MongoConfig struct {
	URI          string `json:"uri" env:"MONGO_URI"`
	User         string `json:"user" env:"MONGO_USER"`
	Password     string `json:"password" env:"MONGO_PASS"`
	DatabaseName string `json:"database_name"`
}

type RedisConfig struct {
	URI      string `json:"uri" env:"REDIS_URI"`
	PoolSize int    `json:"pool_size"`
	Password string `json:"password" env:"REDIS_PASS"`
}

type TokenConfig struct {
	AccessTokenTTL  int    `json:"access_token_ttl"`
	RefreshTokenTTL int    `json:"refresh_token_ttl"`
	SignKey         string `json:"sign_key" env:"TOKEN_CONFIG_SIGN_KEY"`
}

type AiServiceConfig struct {
	Addr string `json:"addr"`
}

type NotificationServiceConfig struct {
	Addr string `json:"addr"`
	From string `json:"from"`
}

type PaymentsConfig struct {
	MerchantID       int    `json:"merchant_id" env:"MERCHANT_ID"`
	MerchantPassword string `json:"merchant_password" env:"MERCHANT_PASSWORD"`
	Currency         string `json:"currency"`
	Language         string `json:"language"`
	ResponseURL      string `json:"response_url"`
	CallbackURL      string `json:"callback_url"`
}

type FeedbacksConfig struct {
	Receiver string `json:"receiver"`
}

type CronConfigs struct {
	ValidatePlanSync string `json:"validate_plan_sync"`
}
