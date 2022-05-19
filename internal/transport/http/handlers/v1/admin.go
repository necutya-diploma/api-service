package v1

import (
	"net/http"
	"time"

	"necutya/faker/internal/domain/domain"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

func (h *Handler) initAdminRoutes(router *mux.Router, adminChain alice.Chain) {
	usersRouter := router.PathPrefix("/admin/").Subrouter()

	usersRouter.Handle("/users-report", adminChain.ThenFunc(h.adminGetUsersReport)).Methods(http.MethodGet)
	usersRouter.Handle("/users/{user_id}/feedbacks", adminChain.ThenFunc(h.feedbacksGetByUser)).Methods(http.MethodGet)
	usersRouter.Handle("/users/{user_id}/feedbacks/{feedback_id}/resolve", adminChain.ThenFunc(h.resolveUserFeedback)).Methods(http.MethodPut, http.MethodPatch, http.MethodOptions)
}

type adminUsersReportResponse struct {
	TotalInfo totalInfoForReportResponse  `json:"total_info"`
	UsersInfo []userInfoForReportResponse `json:"users_info"`
}

type totalInfoForReportResponse struct {
	UsersAmount int64   `json:"users_amount"`
	RPM         int64   `json:"rpm"`
	FPM         int64   `json:"fpm"`
	GPM         float64 `json:"gpm"`
}

type userInfoForReportResponse struct {
	ID        string   `json:"id"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Email     string   `json:"email"`
	Plan      string   `json:"plan"`
	GPM       float64  `json:"gpm"`
	RPM       int64    `json:"rpm"`
	Feedbacks []string `json:"feedbacks"`
}

func (h *Handler) adminGetUsersReport(w http.ResponseWriter, r *http.Request) {
	report, err := h.services.Admin.GetUsersReport(r.Context())
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, convertAdminUsersReportToResponse(report))
}

func convertAdminUsersReportToResponse(report *domain.UsersReport) *adminUsersReportResponse {
	userInfoResponse := make([]userInfoForReportResponse, len(report.UsersInfo))

	for i := range report.UsersInfo {
		userInfoResponse[i] = userInfoForReportResponse(report.UsersInfo[i])
	}

	return &adminUsersReportResponse{
		TotalInfo: totalInfoForReportResponse(report.TotalInfo),
		UsersInfo: userInfoResponse,
	}
}

type feedbackResponse struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	Processed  bool   `json:"processed"`
	CreatedAt  string `json:"created_at"`
	ResolvedAt string `json:"resolved_at"`
}

func (h *Handler) feedbacksGetByUser(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	result, err := h.services.Feedback.GetUserFeedbacks(r.Context(), userID.(string))
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendResponse(w, http.StatusOK, convertFeedbackByUserToResponse(result))
}

func convertFeedbackByUserToResponse(feedbacks []*domain.Feedback) []*feedbackResponse {
	feedbacksResponse := make([]*feedbackResponse, len(feedbacks))

	for i := range feedbacks {
		feedbacksResponse[i] = convertFeedbackToResponse(feedbacks[i])
	}

	return feedbacksResponse
}

func convertFeedbackToResponse(feedback *domain.Feedback) *feedbackResponse {
	var resolvedAt string
	if feedback.ResolvedAt != nil {
		resolvedAt = feedback.ResolvedAt.Format(time.RFC822)
	}
	return &feedbackResponse{
		ID:         feedback.ID.String(),
		Text:       feedback.Text,
		Processed:  feedback.Processed,
		CreatedAt:  feedback.CreatedAt.Format(time.RFC822),
		ResolvedAt: resolvedAt,
	}
}

func (h *Handler) resolveUserFeedback(w http.ResponseWriter, r *http.Request) {
	userID, err := GetPathVar(r, "user_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	feedbackID, err := GetPathVar(r, "feedback_id", stringType)
	if err != nil {
		SendEmptyResponse(w, http.StatusBadRequest)
		return
	}

	err = h.services.Feedback.ResolveFeedback(r.Context(), userID.(string), feedbackID.(string))
	if err != nil {
		SendHTTPError(w, err)
		return
	}

	SendEmptyResponse(w, http.StatusCreated)
}
