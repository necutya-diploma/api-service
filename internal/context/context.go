package context

import (
	"context"

	"github.com/google/uuid"
)

const (
	userIDKey    = "user-id"
	userRoleKey  = "user-role"
	sessionIDKey = "session-id"
	tokenIDKey   = "token-id"
	planIDKey    = "plan-id"
)

// GetUserID - returns user ID from context.
func GetUserID(ctx context.Context) string {
	value, _ := ctx.Value(userIDKey).(string)

	return value
}

// WithUserID - add user ID value to context.
func WithUserID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, userIDKey, value) // nolint
}

// GetTokenID - returns token UUID from context.
func GetRole(ctx context.Context) string {
	value, _ := ctx.Value(userRoleKey).(string)

	return value
}

// WithTokenID - add token UUID value to context.
func WithRole(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, userRoleKey, value) // nolint
}

// GetSessionID - returns session UUID from context.
func GetSessionID(ctx context.Context) uuid.UUID {
	value, _ := ctx.Value(sessionIDKey).(string)

	uuidValue, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}
	}

	return uuidValue
}

// WithSessionID - add session UUID value to context.
func WithSessionID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, sessionIDKey, value) // nolint
}

// GetTokenID - returns token UUID from context.
func GetTokenID(ctx context.Context) uuid.UUID {
	value, _ := ctx.Value(tokenIDKey).(string)

	uuidValue, err := uuid.Parse(value)
	if err != nil {
		return uuid.UUID{}
	}

	return uuidValue
}

// WithTokenID - add token UUID value to context.
func WithTokenID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, tokenIDKey, value) // nolint
}

// GetTokenID - returns token UUID from context.
func GetPlanID(ctx context.Context) string {
	value, _ := ctx.Value(planIDKey).(string)

	return value
}

// WithTokenID - add token UUID value to context.
func WithPlanID(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, planIDKey, value) // nolint
}
