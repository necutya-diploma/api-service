package httpTransport

import (
	"context"
	"net/http"

	"necutya/faker/internal/config"
	"necutya/faker/pkg/logger"

	"golang.org/x/sync/errgroup"
)

type HttpServer struct {
	httpServer *http.Server
	config     *config.HTTPConfig
}

func NewHttp(config *config.HTTPConfig, handler http.Handler) *HttpServer {
	return &HttpServer{
		httpServer: &http.Server{
			Addr:           config.Address,
			Handler:        handler,
			ReadTimeout:    config.ReadTimeout,
			WriteTimeout:   config.WriteTimeout,
			MaxHeaderBytes: config.MaxHeaderMegabytes << 20,
		},
	}
}

func (s *HttpServer) Run(ctx context.Context) {
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		logger.Infof("Starting http server: addr=%v", s.httpServer.Addr)
		return s.httpServer.ListenAndServe()
	})

	g.Go(func() error {
		<-gCtx.Done()
		logger.Info("Stopping http server...")
		return s.httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		logger.Info("Http server has been stopped successfully")
	}
}
