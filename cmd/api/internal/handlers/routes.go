package handler

import (
	"log"
	"net/http"
	"os"

	reader "github.com/igomonov88/nimbler_reader/proto"
	writer "github.com/igomonov88/nimbler_writer/proto"

	"nimbler_gateway/internal/mid"
	"nimbler_gateway/internal/platform/auth"
	"nimbler_gateway/internal/platform/web"
)

type Gateway struct {
	authenticator *auth.Authenticator
	wrt           writer.WriterClient
}

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, wrt writer.WriterClient, rdr reader.ReaderClient) http.Handler {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))
	check := Check{build: build, wrt: wrt, rdr: rdr}

	app.Handle(http.MethodGet, "/v1/health", check.Health)

	return app
}