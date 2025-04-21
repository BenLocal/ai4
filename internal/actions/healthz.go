package actions

import (
	"github.com/benlocal/ai4/internal/db"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Healthz struct {
	db db.Datebase
}

func NewHealthz(db db.Datebase) *Healthz {
	return &Healthz{db: db}
}

func (h *Healthz) AddRouters(router *router.Router) {
	router.GET("/healthz", h.healthzHandler)
}

func (h *Healthz) healthzHandler(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.WriteString("up")
}
