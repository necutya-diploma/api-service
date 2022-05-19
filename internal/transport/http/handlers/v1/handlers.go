package v1

import (
	"net/http"

	"necutya/faker/internal/service"
	"necutya/faker/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
)

const (
	version1 = "/v1"
)

type Handler struct {
	services *service.Service
	router   *mux.Router
	validate *validator.Validate
}

func NewHandler(services *service.Service, router *mux.Router, validate *validator.Validate) *Handler {
	return &Handler{
		services: services,
		router:   router,
		validate: validate,
	}
}

func (h *Handler) Init(URLPrefix, externalURLPrefix string, CORSAllowedHost []string) *mux.Router {
	var (
		router         = h.router
		internalRouter = router.PathPrefix(URLPrefix).Subrouter()
		externalRouter = router.PathPrefix(externalURLPrefix).Subrouter()
	)

	h.initInternal(internalRouter, CORSAllowedHost)

	h.initExternal(externalRouter)

	return router
}

func (h *Handler) initInternal(router *mux.Router, CORSAllowedHost []string) {
	var (
		v1InternalRouter = router.PathPrefix(version1).Subrouter()

		publicChain   = alice.New()
		authChain     = publicChain.Append(h.AuthMiddleware)
		authUserChain = authChain.Append(h.UserMiddleware)

		adminChain = publicChain.Append(h.AuthMiddleware, h.AdminMiddleware)
	)

	v1InternalRouter.Handle("/ping", publicChain.ThenFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("pong")
	})).Methods(http.MethodGet)

	h.initUsersRoutes(v1InternalRouter, publicChain, authUserChain)
	h.initCheckRoutes(v1InternalRouter, authChain)
	h.initAdminRoutes(v1InternalRouter, adminChain)
	h.initPaymentRoutes(v1InternalRouter, publicChain)
	h.initPlansRoutes(v1InternalRouter, publicChain)
	h.initFeedbacksRoutes(v1InternalRouter, publicChain)
}

func (h *Handler) initExternal(router *mux.Router) {
	var (
		v1ExternalRouter = router.PathPrefix(version1).Subrouter()

		externalPrivateChain = alice.New(h.ExternalAuthMiddleware)
	)

	h.initExternalRoutes(v1ExternalRouter, externalPrivateChain)
}
