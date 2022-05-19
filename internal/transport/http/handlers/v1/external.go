package v1

import (
	"net/http"

	reqContext "necutya/faker/internal/context"
	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/service"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initExternalRoutes(router *mux.Router, externalPrivateChain alice.Chain) {
	router.Handle("/check-message", externalPrivateChain.ThenFunc(h.checkMessageExternal)).Methods(http.MethodPost, http.MethodOptions)
}

type checkMessageRequest struct {
	Message  string `json:"message" validate:"required,min=3"`
	SaveToDb bool   `json:"save_to_db"`
}

type checkMessageResponse struct {
	Message          string  `json:"message"`
	IsGenerated      bool    `json:"is_generated"`
	GeneratedPercent float64 `json:"generated_percent"`
}

func (h *Handler) checkMessageExternal(w http.ResponseWriter, r *http.Request) {
	var input checkMessageRequest

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

	ctx := r.Context()

	res, err := h.services.AI.CheckMessage(
		ctx,
		reqContext.GetUserID(ctx),
		reqContext.GetPlanID(ctx),
		&service.CheckMessageInput{
			Message:  input.Message,
			SaveToDb: input.SaveToDb,
		},
		domain.External,
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, checkMessageResponse(*res))
}
