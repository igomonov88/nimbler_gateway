package handler

import (
	"context"
	"net/http"

	srv "github.com/igomonov88/nimbler_server/proto"
	"go.opencensus.io/trace"

	"nimbler_gateway/internal/platform/web"
)

// Check provides support for orchestration health checks.
type Check struct {
	build string
	srv   srv.ServerClient
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Check.Health")
	defer span.End()

	health := struct {
		GatewayVersion     string `json:"gateway_version"`
		UserServiceVersion string `json:"user_service_version"`
		Status             string `json:"status"`
	}{
		GatewayVersion: c.build,
	}

	resp, err := c.srv.HealthCheck(ctx, nil)
	if err != nil {
		// If the one of the depending services is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.

		health.Status = "service is not ready"
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}
	health.Status = "ok"
	health.UserServiceVersion = resp.GetVersion()
	return web.Respond(ctx, w, &health, http.StatusOK)
}
