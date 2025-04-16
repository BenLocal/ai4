package actions

import "github.com/fasthttp/router"

type Actions interface {
	// AddAction adds a new action to the router.
	AddRouters(router *router.Router)
}
