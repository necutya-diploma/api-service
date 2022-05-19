package v1

import (
	"net/http"

	reqContext "necutya/faker/internal/context"
	"necutya/faker/internal/domain/domain"
	"necutya/faker/internal/service"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initCheckRoutes(router *mux.Router, publicChain alice.Chain) {
	router.Handle("/check-message", publicChain.ThenFunc(h.checkMessageInternal)).Methods(http.MethodPost, http.MethodOptions)
}

func (h *Handler) checkMessageInternal(w http.ResponseWriter, r *http.Request) {
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
		domain.Internal,
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, checkMessageResponse(*res))
}
