package handlers

import (
	"net/http"

	"necutya/faker/internal/service"
	v1 "necutya/faker/internal/transport/http/handlers/v1"

	"github.com/go-chi/cors"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
)

type Handler struct {
	services *service.Service
	router   *mux.Router
	validate *validator.Validate
}

func New(services *service.Service, validate *validator.Validate) *Handler {
	handler := &Handler{
		services: services,
		router:   mux.NewRouter(),
		validate: validate,
	}

	return handler
}

func (h *Handler) Init(URLPrefix, externalURLPrefix string, CORSAllowedHost []string) http.Handler {
	v1Router := v1.NewHandler(h.services, h.router, h.validate)

	v1Router.Init(URLPrefix, externalURLPrefix, CORSAllowedHost)

	return getCors().Handler(h.router)
}

func getCors() *cors.Cors {
	return cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{http.MethodHead, http.MethodGet, http.MethodPost, http.MethodPut,
			http.MethodDelete, http.MethodOptions, http.MethodPatch},
		AllowedHeaders:   []string{"Origin", "Authorization", "Content-Type"},
		AllowCredentials: true,
		ExposedHeaders:   []string{""},
		MaxAge:           10,
	})
}
