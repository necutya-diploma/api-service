package v1

import (
	"log"
	"net/http"

	"necutya/faker/internal/context"
	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/service"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

const (
	addNotification = "add"
)

func (h *Handler) initUsersRoutes(router *mux.Router, publicChain, privateChain alice.Chain) {
	usersRouter := router.PathPrefix("/users").Subrouter()

	router.Handle("/sign-up", publicChain.ThenFunc(h.userSignUp)).Methods(http.MethodPost, http.MethodOptions)
	router.Handle("/sign-in", publicChain.ThenFunc(h.userSignIn)).Methods(http.MethodPost, http.MethodOptions)
	usersRouter.Handle("/{user_id}/refresh", publicChain.ThenFunc(h.userRefresh)).Methods(http.MethodPost, http.MethodOptions)

	usersRouter.Handle("/{user_email}/confirm", publicChain.ThenFunc(h.userConfirmEmail)).Methods(http.MethodPost, http.MethodOptions)

	usersRouter.Handle("/{user_email}/password-reset", publicChain.ThenFunc(h.passwordResetRequest)).Methods(http.MethodGet)
	usersRouter.Handle("/{user_email}/password-reset", publicChain.ThenFunc(h.passwordReset)).Methods(http.MethodPost, http.MethodOptions)
	usersRouter.Handle("/{user_email}/password-reset/confirm", publicChain.ThenFunc(h.verifyPasswordReset)).Methods(http.MethodPost, http.MethodOptions)

	usersRouter.Handle("/{user_id}/sign-out", privateChain.ThenFunc(h.userSignOut)).Methods(http.MethodDelete, http.MethodOptions)

	usersRouter.Handle("/{user_id}", privateChain.ThenFunc(h.userGetOne)).Methods(http.MethodGet)
	usersRouter.Handle("/{user_id}", privateChain.ThenFunc(h.userUpdate)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
	usersRouter.Handle("/{user_id}/update-password", privateChain.ThenFunc(h.userUpdatePassword)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
	usersRouter.Handle("/{user_id}", privateChain.ThenFunc(h.userDelete)).Methods(http.MethodDelete)

	usersRouter.Handle("/{user_id}/notification/{action:\\b*add|remove\\b*}", privateChain.ThenFunc(h.userUpdateNotification)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
	usersRouter.Handle("/{user_id}/plan/update", privateChain.ThenFunc(h.userUpdatePlan)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
	usersRouter.Handle("/{user_id}/external/regenerate", privateChain.ThenFunc(h.userRegenerateExternal)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
}

type userSignUpRequest struct {
	FirstName           string `json:"first_name" validate:"required"`
	LastName            string `json:"last_name" validate:"required"`
	Email               string `json:"email" validate:"required,email,max=64"`
	Password            string `json:"password" validate:"required,min=8,max=64"`
	ReceiveNotification bool   `json:"receive_notification"`
}

type userSignUpResponse struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email,max=64"`
}

func (h *Handler) userSignUp(w http.ResponseWriter, r *http.Request) {
	var input userSignUpRequest

	err := UnmarshalRequest(r, &input)
	if err != nil {
		log.Println(err)
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	user, err := h.services.Auth.SignUp(r.Context(), service.UserSignUpInput{
		FirstName:           input.FirstName,
		LastName:            input.LastName,
		Email:               input.Email,
		Password:            input.Password,
		ReceiveNotification: input.ReceiveNotification,
	})
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, userSignUpResponse{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	})
}

type userSignInRequest struct {
	Email    string `json:"email" validate:"required,email,max=64"`
	Password string `json:"password" validate:"required,max=64"`
}

type userSignInResponse struct {
	Tokens userSignInTokens `json:"tokens"`
	User   *userResponse    `json:"user"`
}

type userSignInTokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) userSignIn(w http.ResponseWriter, r *http.Request) {
	var input userSignInRequest

	err := UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	tokens, user, err := h.services.Auth.SignIn(r.Context(), service.UserSignInInput{
		Email:     input.Email,
		Password:  input.Password,
		Client:    GetUserClient(r),
		IPAddress: GetUserIPAddress(r),
	})
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, userSignInResponse{
		Tokens: userSignInTokens(*tokens),
		User:   convertUserToUserResponse(user),
	})
}

type userConfirmEmailRequest struct {
	Code string `json:"code" validate:"required,min=8"`
}

func (h *Handler) userConfirmEmail(w http.ResponseWriter, r *http.Request) {
	var input userConfirmEmailRequest
	log.Println("here")
	email, err := GetPathVar(r, "user_email", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	err = h.services.Auth.ConfirmUserEmail(r.Context(), email.(string), input.Code)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}

type userRefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required,min=32"`
}

type userRefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) userRefresh(w http.ResponseWriter, r *http.Request) {
	var input userRefreshTokenRequest

	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	requestContext := r.Context()

	tokens, err := h.services.Auth.RefreshToken(
		requestContext,
		userID.(string),
		input.RefreshToken,
	)
	if err != nil {
		SendEmptyResponse(w, http.StatusNotAcceptable)
		return
	}

	SendResponse(w, http.StatusCreated, userRefreshTokenResponse(*tokens))
}

func (h *Handler) userSignOut(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	requestContext := r.Context()

	err = h.services.Auth.SignOut(
		requestContext,
		userID.(string),
		context.GetTokenID(requestContext),
		context.GetSessionID(requestContext),
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusNoContent)
}

type userResponse struct {
	ID            string `json:"id"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	PlanID        string `json:"plan_id"`
	ExternalToken string `json:"external_token"`
	Role          string `json:"role"`

	TodayInternalRequest int `json:"today_internal_request"`
	TodayExternalRequest int `json:"today_external_request"`

	ReceiveNotification bool `json:"receive_notification"`
}

func convertUserToUserResponse(user *domain.User) *userResponse {
	return &userResponse{
		ID:                   user.ID,
		Email:                user.Email,
		LastName:             user.LastName,
		FirstName:            user.FirstName,
		PlanID:               user.PlanID,
		ReceiveNotification:  user.ReceiveNotification,
		ExternalToken:        user.Credential.ExternalToken,
		Role:                 string(user.Role),
		TodayInternalRequest: user.TodayInternalRequest,
		TodayExternalRequest: user.TodayExternalRequest,
	}
}

func (h *Handler) userGetOne(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	user, err := h.services.User.GetOne(
		r.Context(),
		userID.(string),
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, convertUserToUserResponse(user))
}

type userUpdateRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func (h *Handler) userUpdate(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	var input userUpdateRequest

	err = UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	user, err := h.services.User.Update(
		r.Context(),
		userID.(string),
		service.UserUpdateInput{
			FirstName: input.FirstName,
			LastName:  input.LastName,
		},
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, convertUserToUserResponse(user))
}

func (h *Handler) userDelete(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.User.Delete(
		r.Context(),
		userID.(string),
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusNoContent)
}

func (h *Handler) userUpdateNotification(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	action, err := GetPathVar(r, "action", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	user, err := h.services.User.UpdateNotification(
		r.Context(),
		userID.(string),
		action == addNotification,
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, convertUserToUserResponse(user))
}

func (h *Handler) passwordResetRequest(w http.ResponseWriter, r *http.Request) {
	email, err := GetPathVar(r, "user_email", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.Auth.PasswordResetRequest(r.Context(), email.(string))
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusOK)
}

type userVerifyPasswordResetRequest struct {
	Code string `json:"code" validate:"required,min=8"`
}

func (h *Handler) verifyPasswordReset(w http.ResponseWriter, r *http.Request) {
	var input userVerifyPasswordResetRequest

	err := UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	email, err := GetPathVar(r, "user_email", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.Auth.VerifyPasswordReset(r.Context(), email.(string), input.Code)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}

type userPasswordReset struct {
	Password        string `json:"password" validate:"required,min=8,max=64"`
	PasswordConfirm string `json:"password_confirm" validate:"required,min=8,max=64,eqfield=Password"`
}

func (h *Handler) passwordReset(w http.ResponseWriter, r *http.Request) {
	var input userPasswordReset

	err := UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	email, err := GetPathVar(r, "user_email", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.Auth.PasswordReset(r.Context(), email.(string), service.UserPasswordResetInput{
		Password:        input.Password,
		PasswordConfirm: input.PasswordConfirm,
	})
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}

type userUpdatePasswordRequest struct {
	CurrentPassword    string `json:"current_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8,max=64,nefield=CurrentPassword"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required,min=8,max=64,eqfield=NewPassword"`
}

func (h *Handler) userUpdatePassword(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	var input userUpdatePasswordRequest

	err = UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	err = h.services.User.UpdateUserPassword(
		r.Context(),
		userID.(string),
		service.UserUpdatePasswordInput{
			CurrentPassword: input.CurrentPassword,
			NewPassword:     input.NewPassword,
		},
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}

type userUpdatePlanRequest struct {
	PlanID string `json:"plan_id" validate:"required"`
}

type userUpdatePlanResponse struct {
	CheckoutLink string `json:"checkout_link"`
}

func (h *Handler) userUpdatePlan(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	var input userUpdatePlanRequest

	err = UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.validate.Struct(input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	checkoutLink, err := h.services.User.UpdateUserPlan(
		r.Context(),
		userID.(string),
		input.PlanID,
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, userUpdatePlanResponse{CheckoutLink: checkoutLink})
}

func (h *Handler) userRegenerateExternal(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	user, err := h.services.User.RegenerateExternalToken(
		r.Context(),
		userID.(string),
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusCreated, convertUserToUserResponse(user))
}
