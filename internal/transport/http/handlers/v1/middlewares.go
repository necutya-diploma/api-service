package v1

import (
	"net/http"

	reqContext "necutya/faker/internal/context"
)

// AuthMiddleware - authenticate User by JWT token and add to context his ID and session ID.
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			authToken = GetAuthorizationHeader(r)
		)

		accessToken, err := ParseBearerAuthorizationHeader(authToken)
		if err != nil {
			SendEmptyResponse(w, http.StatusUnauthorized)
			return
		}

		tokenInfo, err := h.services.Auth.ValidateToken(r.Context(), accessToken)
		if err != nil {
			SendEmptyResponse(w, http.StatusUnauthorized)
			return
		}

		ctx := reqContext.WithUserID(r.Context(), tokenInfo.UserID)
		ctx = reqContext.WithPlanID(ctx, tokenInfo.PlanID)
		ctx = reqContext.WithSessionID(ctx, tokenInfo.SessionID)
		ctx = reqContext.WithTokenID(ctx, tokenInfo.TokenID)

		SetUserIDHeader(r, tokenInfo.UserID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) UserMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := GetPathVar(r, "user_id", stringType)
		if err != nil {
			SendEmptyResponse(w, http.StatusBadRequest)
			return
		}

		userContextID := reqContext.GetUserID(r.Context())
		if userID != userContextID {
			SendEmptyResponse(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := reqContext.GetUserID(r.Context())
		user, err := h.services.User.GetOne(r.Context(), userID)
		if err != nil {
			SendEmptyResponse(w, http.StatusBadRequest)
			return
		}

		if !user.IsAdmin() {
			SendEmptyResponse(w, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// ExternalAuthMiddleware - authenticate User by external token.
func (h *Handler) ExternalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			externalToken = GetExternalAuthorizationHeader(r)
		)

		userID, err := h.services.Auth.ValidateExternalToken(r.Context(), externalToken)
		if err != nil {
			SendHTTPError(w, err)
			return
		}

		ctx := reqContext.WithUserID(r.Context(), userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
