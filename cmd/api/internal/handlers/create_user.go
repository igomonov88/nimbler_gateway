package handler

import (
	"context"
	"net/http"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) CreateUser(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.CreateUser")
	defer span.End()

	var cur CreateUserRequest
	if err := web.Decode(r, &cur); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	req := writer.CreateUserRequest{
		Name:     cur.Name,
		Email:    cur.Email,
		Password: cur.Password,
	}

	nu, err := g.wrt.CreateUser(ctx, &req)
	if err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}
	resp := CreateUserResponse{UserID: nu.GetUserID()}

	return web.Respond(ctx, w, &resp, http.StatusOK)
}
