package v1

import (
	"net/http"

	"necutya/faker/internal/domain/domain"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initPlansRoutes(router *mux.Router, publicChain alice.Chain) {
	plansRouter := router.PathPrefix("/plans").Subrouter()

	plansRouter.Handle("", publicChain.ThenFunc(h.getPlans)).Methods(http.MethodGet)
	plansRouter.Handle("/{plan_id}", publicChain.ThenFunc(h.getPlan)).Methods(http.MethodGet)
}

type planResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	Options              []string `json:"options"`
	Price                int      `json:"price"`
	Duration             int      `json:"duration"`
	InternalRequestCount int      `json:"internal_request_count"`
	ExternalRequestCount int      `json:"external_request_count"`
}

func convertPlansToPlansManyResponse(plans []*domain.Plan) []*planResponse {
	planResponses := make([]*planResponse, len(plans))

	for i := range plans {
		planResponses[i] = convertPlanToPlanResponse(plans[i])
	}

	return planResponses
}

func convertPlanToPlanResponse(plan *domain.Plan) *planResponse {
	return &planResponse{
		ID:                   plan.ID,
		Name:                 plan.Name,
		Description:          plan.Description,
		Options:              plan.Options,
		Price:                plan.Price,
		Duration:             plan.Duration,
		InternalRequestCount: plan.InternalRequestsCount,
		ExternalRequestCount: plan.ExternalRequestsCount,
	}
}

func (h *Handler) getPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.services.Plan.GetMany(r.Context())
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, convertPlansToPlansManyResponse(plans))
}

func (h *Handler) getPlan(w http.ResponseWriter, r *http.Request) {
	planID, err := GetPathVar(r, "plan_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	plan, err := h.services.Plan.GetOne(
		r.Context(),
		planID.(string),
	)
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, convertPlanToPlanResponse(plan))
}
