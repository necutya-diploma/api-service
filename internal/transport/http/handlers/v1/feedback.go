package v1

import (
	"log"
	"net/http"

	"necutya/faker/internal/service"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initFeedbacksRoutes(router *mux.Router, chain alice.Chain) {
	usersRouter := router.PathPrefix("/feedbacks").Subrouter()

	usersRouter.Handle("", chain.ThenFunc(h.createFeedback)).Methods(http.MethodPost, http.MethodOptions)
}

type createFeedbackRequest struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
	Text     string `json:"text" validate:"required"`
}

func (h *Handler) createFeedback(w http.ResponseWriter, r *http.Request) {
	var input createFeedbackRequest

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

	err = h.services.Feedback.SetFeedback(
		r.Context(),
		service.FeedbackCreateInput{
			UserID:   input.UserID,
			Email:    input.Email,
			Username: input.Username,
			Text:     input.Text,
		},
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}
