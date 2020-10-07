package main

import (
	"context"
	"expvar"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"

	"contrib.go.opencensus.io/exporter/zipkin"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"

	srv "github.com/igomonov88/nimbler_server/proto"

	handler "nimbler_gateway/cmd/internal/handlers"
	"nimbler_gateway/config"
)

// build is the git version of this program. It is set using build flags in the
// makefile.
var build = "develop"

func main() {
	if err := run(); err != nil {
		log.Println("error :", err)
		os.Exit(1)
	}
}

func run() error {
	// =====================  Logging  ========================

	log := log.New(os.Stdout, "GATEWAY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// ================== Read Configuration ==================

	cfg, err := config.Parse("config/config.yaml")
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}

	// ======================  App Starting ====================

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application initializing : version %q", build)
	defer log.Println("main : Completed")

	log.Println("main : Started : Initializing  support")

	// ============= Connect to appropriate servers ============

	conn, err := grpc.DialContext(context.Background(), cfg.Server.APIHost, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to create grpc connection")
	}
	srvClient := srv.NewServerClient(conn)

	defer func() {
		log.Printf("main : Server Connection Stopping : %s", cfg.Server.APIHost)
		conn.Close()
	}()

	// ================= Start Tracing Support =================

	log.Println("main : Started : Initializing zipkin tracing support")

	localEndpoint, err := openzipkin.NewEndpoint(cfg.Zipkin.ServiceName, cfg.Zipkin.LocalEndpoint)
	if err != nil {
		return err
	}

	reporter := zipkinHTTP.NewReporter(cfg.Zipkin.ReporterURI)
	ze := zipkin.NewExporter(reporter, localEndpoint)

	trace.RegisterExporter(ze)
	trace.ApplyConfig(trace.Config{
		DefaultSampler: trace.ProbabilitySampler(cfg.Zipkin.Probability),
	})

	defer func() {
		log.Printf("main : Tracing Stopping : %s", cfg.Zipkin.LocalEndpoint)
		reporter.Close()
	}()

	// =================== Start Debug Service ===================
	//
	// /debug/pprof - Added to the default mux by importing the net/http/pprof package.
	// /debug/vars - Added to the default mux by importing the expvar package.
	//
	// Not concerned with shutting this down when the application is shutdown.

	log.Println("main : Started : Initializing debugging support")

	go func() {
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", http.ListenAndServe(cfg.Web.DebugHost, http.DefaultServeMux))
	}()

	// ==================== Start API Service  =====================

	log.Println("main : Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handler.API(build, shutdown, log, srvClient),
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// ====================  Shutdown  =============================

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return errors.Wrap(err, "server error")

	case sig := <-shutdown:
		log.Printf("main : %v : Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			err = api.Close()
		}

		// Log the status of this shutdown.
		switch {
		case sig == syscall.SIGSTOP:
			return errors.New("integrity issue caused shutdown")
		case err != nil:
			return errors.Wrap(err, "could not stop server gracefully")
		}
	}

	return nil
}
