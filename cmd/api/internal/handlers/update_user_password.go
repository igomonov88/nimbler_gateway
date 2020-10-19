package handler

import (
	"context"
	"net/http"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) UpdateUserPassword(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.UpdateUserPassword")
	defer span.End()

	var uup UpdateUserPasswordRequest
	if err := web.Decode(r, &uup); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	req := writer.UpdateUserPasswordRequest{
		UserID: uup.UserID,
		Password: uup.Password,
	}

	_, err := g.wrt.UpdateUserPassword(ctx, &req)
	if err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, struct{}{}, http.StatusOK)
}

