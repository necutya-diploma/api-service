package core

import "errors"

var (
	ErrAlreadyExist = errors.New("already exists")
	ErrNotFound     = errors.New("not found")

	ErrInvalidLoginOrPassword = errors.New("invalid login or password")
	ErrExpiredSession         = errors.New("session is expired")
	ErrInvalidSession         = errors.New("session is invalid")
	ErrUnconfirmedEmail       = errors.New("unconfirmed email")

	ErrExpiredCode = errors.New("code is expired, try one more time")
	ErrInvalidCode = errors.New("invalid verification code")

	ErrUnverifiedPasswordReset = errors.New("unverified password reset")
	ErrInvalidCurrentPassword  = errors.New("invalid current password")

	ErrRequestLimit = errors.New("requests limit for today has been reached")

	ErrUnknownCallbackType = errors.New("unknown callback type")
	ErrTransactionInvalid  = errors.New("invalid transaction")

	ErrThisPlanAlreadySet = errors.New("this plan already set")
)

type ApiError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}
