package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httprouter "pubsub-chat/internal/http"
	"pubsub-chat/internal/hub"
	"pubsub-chat/internal/logger"
	"pubsub-chat/pkg/version"
)

func main() {
	lg := logger.New()
	defer lg.Sync()

	addr := getEnv("HTTP_ADDR", ":8080")

	// Start Hub
	h := hub.New(lg)
	go h.Run()

	r := httprouter.New(h, lg)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	lg.Infow("server.start", "addr", addr, "version", version.String())
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			lg.Errorw("server.listen", "err", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	h.Stop()
	if err := srv.Shutdown(ctx); err != nil {
		lg.Errorw("server.shutdown", "err", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
