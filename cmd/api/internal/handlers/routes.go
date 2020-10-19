package handler

import (
	"log"
	"net/http"
	"os"

	reader "github.com/igomonov88/nimbler_reader/proto"
	writer "github.com/igomonov88/nimbler_writer/proto"

	"github.com/igomonov88/nimbler_gateway/internal/mid"
	"github.com/igomonov88/nimbler_gateway/internal/platform/auth"
	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

type Gateway struct {
	authenticator *auth.Authenticator
	wrt           writer.WriterClient
}

// API constructs an http.Handler with all application routes defined.
func API(build string, shutdown chan os.Signal, log *log.Logger, wrt writer.WriterClient, rdr reader.ReaderClient) http.Handler {
	app := web.NewApp(shutdown, mid.Logger(log), mid.Errors(log), mid.Metrics(), mid.Panics(log))
	check := Check{build: build, wrt: wrt, rdr: rdr}
	gateway := Gateway{wrt: wrt}

	app.Handle(http.MethodGet, "/v1/health", check.Health)
	app.Handle(http.MethodPost, "/v1/create_user", gateway.CreateUser)
	app.Handle(http.MethodGet, "/v1/user/:user_id", gateway.RetrieveUser)
	app.Handle(http.MethodDelete, "/v1/user/:user_id", gateway.DeleteUser)
	app.Handle(http.MethodPost, "/v1/update_user_info", gateway.UpdateUserInfo)
	app.Handle(http.MethodPost, "/v1/update_user_password", gateway.UpdateUserPassword)
	app.Handle(http.MethodPost, "/v1/does_email_exist", gateway.DoesEmailExists)
	return app
}
