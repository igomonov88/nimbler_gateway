package handler

import (
	"log"
	"net/http"
	"os"

	srv "github.com/igomonov88/nimbler_server/proto"

	"nimbler_gateway/internal/mid"
	"nimbler_gateway/internal/platform/auth"
	"nimbler_gateway/internal/platform/web"
)

type Gateway struct {
	authenticator *auth.Authenticator
	srv           srv.ServerClient
}

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, srv srv.ServerClient) http.Handler {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))

	check := Check{
		build: build,
		srv:   srv,
	}

	app.Handle(http.MethodGet, "/v1/health", check.Health)

	return app
}
