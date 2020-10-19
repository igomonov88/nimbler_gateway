package handler

import (
	"context"
	"net/http"
	"strings"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"go.opencensus.io/trace"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) DeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.DeleteUser")
	defer span.End()

	userID := strings.TrimSpace(params["user_id"])

	req := writer.DeleteUserRequest{UserID: userID}

	_, err := g.wrt.DeleteUser(ctx, &req)
	if err != nil {
		return web.NewRequestError(err, http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, struct{}{}, http.StatusOK)
}

