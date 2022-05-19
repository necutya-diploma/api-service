package token_manager

import (
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const refreshTokenLength = 32

var (
	ErrInvalidSignMethod = errors.New("unexpected signing method")
	ErrInvalidClaims     = errors.New("invalid token claims")
	ErrTokenExpired      = errors.New("token is expired")
)

type JWTManager struct {
	signKey string
}
type Claims struct {
	SessionID string
	TokenID   string
	UserID    string
	PlanID    string

	jwt.StandardClaims
}

func NewJWT(signKey string) (JWTManager, error) {
	if signKey == "" {
		return JWTManager{}, errors.New("empty signing key")
	}

	return JWTManager{signKey: signKey}, nil
}

func (tm JWTManager) GenerateAccessToken(userID, planID, sessionID, tokenID string, ttl int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * time.Duration(ttl)).Unix(),
		},
		UserID:    userID,
		SessionID: sessionID,
		TokenID:   tokenID,
		PlanID:    planID,
	})

	return token.SignedString([]byte(tm.signKey))
}

func (tm JWTManager) GenerateRefreshToken() (string, error) {
	return generateRandomString(refreshTokenLength)
}

func (tm JWTManager) Parse(accessToken string) (map[string]interface{}, error) {
	token, err := jwt.ParseWithClaims(accessToken, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSignMethod
		}

		return []byte(tm.signKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	claimsMap := map[string]interface{}{
		"session_id": claims.SessionID,
		"user_id":    claims.UserID,
		"token_id":   claims.TokenID,
		"plan_id":    claims.PlanID,
	}

	return claimsMap, nil
}
