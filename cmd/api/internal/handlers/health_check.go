package handler

import (
	"context"
	"net/http"

	reader "github.com/igomonov88/nimbler_reader/proto"
	writer "github.com/igomonov88/nimbler_writer/proto"
	"go.opencensus.io/trace"

	"nimbler_gateway/internal/platform/web"
)

// Check provides support for orchestration health checks.
type Check struct {
	build string
	wrt   writer.WriterClient
	rdr   reader.ReaderClient
}

// Health validates the service is healthy and ready to accept requests.
func (c *Check) Health(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	ctx, span := trace.StartSpan(ctx, "handlers.Check.Health")
	defer span.End()

	health := struct {
		GatewayVersion       string `json:"gateway_version"`
		WriterServiceVersion string `json:"writer_service_version"`
		ReaderServiceVersion string `json:"reader_service_version"`
		Status               string `json:"status"`
	}{
		GatewayVersion: c.build,
	}
	wrtResp, err := c.wrt.HealthCheck(ctx, &writer.HealthCheckRequest{})
	if err != nil {
		// If the one of the depending services is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.

		health.Status = err.Error()
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	rdrRest, err := c.rdr.HealthCheck(ctx, &reader.HealthCheckRequest{})
	if err != nil {
		// If the one of the depending services is not ready we will tell the client and use a 500
		// status. Do not respond by just returning an error because further up in
		// the call stack will interpret that as an unhandled error.

		health.Status = err.Error()
		return web.Respond(ctx, w, health, http.StatusInternalServerError)
	}

	health.Status = "ok"
	health.WriterServiceVersion = wrtResp.GetVersion()
	health.ReaderServiceVersion = rdrRest.GetVersion()
	return web.Respond(ctx, w, &health, http.StatusOK)
}
