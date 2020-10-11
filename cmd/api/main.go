package main

import (
	"context"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"contrib.go.opencensus.io/exporter/zipkin"
	"github.com/ardanlabs/conf"
	reader "github.com/igomonov88/nimbler_reader/proto"
	writer "github.com/igomonov88/nimbler_writer/proto"
	openzipkin "github.com/openzipkin/zipkin-go"
	zipkinHTTP "github.com/openzipkin/zipkin-go/reporter/http"
	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"google.golang.org/grpc"

	handlers "nimbler_gateway/cmd/api/internal/handlers"
)

/*
ZipKin: http://localhost:9411
AddLoad: hey -m GET -c 10 -n 10000 "http://localhost:3000/v1/create_url"
expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,mem:memstats.Alloc"
*/

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down.
Service should start even without a DB running yet.
symbols in profiles: https://github.com/golang/go/issues/23376 / https://github.com/google/pprof/pull/366
*/

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
	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "GATEWAY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration
	var cfg struct {
		Web struct {
			APIHost         string        `conf:"default:0.0.0.0:3000"`
			DebugHost       string        `conf:"default:0.0.0.0:4000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		Zipkin struct {
			LocalEndpoint string  `conf:"default:0.0.0.0:3000"`
			ReporterURI   string  `conf:"default:http://zipkin:9411/api/v2/spans"`
			ServiceName   string  `conf:"default:gateway"`
			Probability   float64 `conf:"default:0.05"`
		}
		ReaderServer struct {
			APIHost         string        `conf:"default:0.0.0.0:6000"`
			DebugHost       string        `conf:"default:0.0.0.0:7000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
		WriterServer struct {
			APIHost         string        `conf:"default:0.0.0.0:7000"`
			DebugHost       string        `conf:"default:0.0.0.0:8000"`
			ReadTimeout     time.Duration `conf:"default:5s"`
			WriteTimeout    time.Duration `conf:"default:5s"`
			ShutdownTimeout time.Duration `conf:"default:5s"`
		}
	}

	if err := conf.Parse(os.Args[1:], "GATEWAY", &cfg); err != nil {
		if err == conf.ErrHelpWanted {
			usage, err := conf.Usage("GATEWAY", &cfg)
			if err != nil {
				return errors.Wrap(err, "generating config usage")
			}
			fmt.Println(usage)
			return nil
		}
		return errors.Wrap(err, "parsing config")
	}

	// =========================================================================
	// App Starting

	// Print the build version for our logs. Also expose it under /debug/vars.
	expvar.NewString("build").Set(build)
	log.Printf("main : Started : Application initializing : version %q", build)
	defer log.Println("main : Completed")

	out, err := conf.String(&cfg)
	if err != nil {
		return errors.Wrap(err, "generating config for output")
	}
	log.Printf("main : Config :\n%v\n", out)

	// =========================================================================
	// Start Tracing Support

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

	// =========================================================================
	// Start ReaderServer Connection
	log.Println("main : Started : Initializing server grpc connection")
	readerConn, err := grpc.Dial(cfg.ReaderServer.APIHost, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to connect to grpc reader server")
	}
	writerConn, err := grpc.Dial(cfg.WriterServer.APIHost, grpc.WithInsecure())
	if err != nil {
		return errors.Wrap(err, "failed to connect to grpc writer server")
	}

	writerClient := writer.NewWriterClient(readerConn)
	readerClient := reader.NewReaderClient(writerConn)
	defer func() {
		log.Printf("main : GRPC Connection Stopping : %s", cfg.ReaderServer.APIHost)
		readerConn.Close()
	}()
	// =========================================================================
	// Start Debug Service
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

	// =========================================================================
	// Start API Service

	log.Println("main : Started : Initializing API support")

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      handlers.API(build, shutdown, log, writerClient, readerClient),
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

	// =========================================================================
	// Shutdown

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
