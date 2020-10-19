package handler

import (
	"context"
	"net/http"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) DoesEmailExists(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.DoesEmailExists")
	defer span.End()

	var d DoesEmailExistsRequest
	if err := web.Decode(r, &d); err != nil {
		return errors.Wrapf(err, "unable to decode payload")
	}

	req := writer.DoesEmailExistRequest{
		Email:    d.Email,
	}

	der, err := g.wrt.DoesEmailExist(ctx, &req)
	if err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}
	resp := DoesEmailExistsResponse{Exist: der.GetExist()}

	return web.Respond(ctx, w, &resp, http.StatusOK)
}
