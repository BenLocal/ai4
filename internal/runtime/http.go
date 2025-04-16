package runtime

import (
	"context"
	"fmt"
	"log"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type HttpServer struct {
	port   int
	router *router.Router

	server *fasthttp.Server
}

func NewHttpServer(port int, router *router.Router) *HttpServer {

	return &HttpServer{
		port:   port,
		router: router,
	}
}

func (h *HttpServer) Name() string {
	return "HttpServer"
}

func (h *HttpServer) Start(ctx context.Context) error {
	h.server = &fasthttp.Server{
		Handler: h.router.Handler,
	}

	address := fmt.Sprintf(":%d", h.port)
	if err := h.server.ListenAndServe(address); err != nil {
		log.Fatalf("Error in ListenAndServe: %v", err)
	}

	return nil
}

func (h *HttpServer) Shutdown() error {
	if h.server != nil {
		if err := h.server.Shutdown(); err != nil {
			log.Printf("Error shutting down server: %v", err)
			return err
		}
	}
	return nil
}
