package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
)

const port = "8080"
const envDataFolder = "DATA_FOLDER"
const envLogDir = "LOG_DIR"

func run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	c, err := newClient()
	if err != nil {
		panic(err)
	}

	dataFolder := os.Getenv(envDataFolder)
	if dataFolder == "" {
		dataFolder = "/data"
	}
	groupsOfQueries, err := loadQueries(dataFolder)
	if err != nil {
		panic(err)
	}

OUTER:
	for {
		log15.Info("new iteration")
		for group, queries := range groupsOfQueries {
			log15.Info("new group", "group", group)
			for _, query := range queries {
				func(ctx context.Context, query string) {
					_, m, err := c.search(ctx, query)
					if err != nil {
						log15.Error(err.Error())
					}
					log15.Info("metrics", "group", group, "query", query, "duration_ms", m.took)
					durationSearchHistogram.WithLabelValues(group).Observe(float64(m.took))
				}(ctx, query)
			}
		}
		select {
		case <-ctx.Done():
			break OUTER
		case <-time.After(600 * time.Second):
		}
	}
}

func startServer(wg *sync.WaitGroup) *http.Server {

	http.HandleFunc("/health", health)
	http.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{Addr: ":" + port}

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			panic(err.Error())
		}
	}()
	return srv
}

func main() {

	logDir := os.Getenv(envLogDir)
	if logDir == "" {
		logDir = "."
	}

	log15.Root().SetHandler(log15.MultiHandler(
		log15.StreamHandler(os.Stderr, log15.LogfmtFormat()),
		log15.StreamHandler(&lumberjack.Logger{
			Filename: filepath.Join(logDir, "search_blitz.log"),
			MaxSize:  10, // Megabyte
			MaxAge:   90, // days
		}, log15.JsonFormat())))

	ctx, cleanup := SignalSensitiveContext()
	defer cleanup()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go run(ctx, &wg)

	wg.Add(1)
	srv := startServer(&wg)
	log15.Info("server running on :" + port)

	<-ctx.Done()
	_ = srv.Shutdown(ctx)
	log15.Info("server shut down gracefully")

	wg.Wait()
}

// SignalSensitiveContext returns a background context that is canceled after receiving an
// interrupt or terminate signal. A second signal will abort the program. This function returns
// the context and a function that should be  deferred by the caller to clean up internal channels.
func SignalSensitiveContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		i := 0
		for range signals {
			cancel()

			if i > 0 {
				os.Exit(1)
			}
			i++
		}
	}()

	return ctx, func() {
		cancel()
		signal.Reset(syscall.SIGINT, syscall.SIGTERM)
		close(signals)
	}
}
