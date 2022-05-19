package v1

import (
	"net/http"

	"necutya/faker/pkg/payments/fondy"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initPaymentRoutes(router *mux.Router, publicChain alice.Chain) {
	paymentsRouter := router.PathPrefix("/payments").Subrouter()
	paymentsRouter.Handle("/subscribe/callback", publicChain.ThenFunc(h.paymentsFondyCallback)).Methods(http.MethodPost, http.MethodOptions)
}

func (h *Handler) paymentsFondyCallback(w http.ResponseWriter, r *http.Request) {
	var input fondy.Callback

	err := UnmarshalRequest(r, &input)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.Payments.ProcessCallback(r.Context(), input)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusOK)
}
