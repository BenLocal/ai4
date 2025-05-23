package runtime

import (
	"context"

	"github.com/benlocal/ai4/internal/actions"
	"github.com/benlocal/ai4/internal/db"
	"github.com/benlocal/ai4/pkg/gracefulshutdown"
	"github.com/fasthttp/router"
)

func Start() error {
	g := gracefulshutdown.New()
	g.CatchSignals()

	db, err := db.DatebaseFactory("sqlite", "file:./app.db")
	if err != nil {
		return err
	}

	healthz := actions.NewHealthz(db)
	models := actions.NewModels(db)
	chats := actions.NewChats(db)

	// Initialize the router
	router := router.New()
	healthz.AddRouters(router)
	models.AddRouters(router)
	chats.AddRouters(router)

	g.Add(NewHttpServer(7080, router))

	ctx := context.Background()
	return g.Start(ctx)
}
