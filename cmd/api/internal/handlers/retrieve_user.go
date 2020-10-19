package handler

import (
	"context"
	"net/http"
	"strings"

	writer "github.com/igomonov88/nimbler_writer/proto"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/status"

	"github.com/igomonov88/nimbler_gateway/internal/platform/web"
)

func (g *Gateway) RetrieveUser(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Gateway.RetrieveUser")
	defer span.End()

	userID := strings.TrimSpace(params["user_id"])

	req := writer.RetrieveUserRequest{UserID: userID}

	ru, err := g.wrt.RetrieveUser(ctx, &req)
	if err != nil {
		status := status.Convert(err)
		switch status.Code() {
		case http.StatusNotFound:
			return web.Respond(ctx, w, struct{}{}, http.StatusNotFound)
		case http.StatusBadRequest:
			return web.Respond(ctx, w, struct{}{}, http.StatusBadRequest)
		default:
			return web.NewRequestError(err, http.StatusInternalServerError)
		}
	}

	resp := RetrieveUserResponse{
		UserID: ru.GetUserID(),
		Name:   ru.GetName(),
		Email:  ru.GetEmail(),
	}

	return web.Respond(ctx, w, &resp, http.StatusOK)
}
