package handler

import (
	"context"
	"net/http"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) UpdateUserInfo(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.UpdateUserInfo")
	defer span.End()

	var uur UpdateUserInfoRequest
	if err := web.Decode(r, &uur); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	req := writer.UpdateUserInfoRequest{
		UserID: uur.UserID,
		Name:     uur.Name,
		Email:    uur.Email,
	}

	_, err := g.wrt.UpdateUserInfo(ctx, &req)
	if err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, struct{}{}, http.StatusOK)
}
