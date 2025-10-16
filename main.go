package main

import (
	"context"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"URL_shortener/internal/config"
	"URL_shortener/internal/httpapi"
	"URL_shortener/internal/middleware"
	"URL_shortener/internal/shortener"
	"URL_shortener/internal/storage"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//go:embed web/*
var embeddedWeb embed.FS

func main() {
	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	store := storage.NewMemoryStorage()
	svc := shortener.NewService(store)
	api := httpapi.NewHandler(svc, cfg.BaseURL)

	mux := http.NewServeMux()
	webFS, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		logger.Error("embed sub fs error", "error", err)
		os.Exit(1)
	}
	staticFS, err := fs.Sub(webFS, "static")
	if err != nil {
		logger.Error("embed static fs error", "error", err)
		os.Exit(1)
	}
	fsys := http.FS(webFS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.FileServer(fsys).ServeHTTP(w, r)
			return
		}
		data, err := fs.ReadFile(webFS, "index.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})
	mux.Handle("/api/", api.Routes())
	mux.Handle("/s/", api.Routes())

	handler := middleware.CORS()(
		middleware.RateLimit(cfg.RateLimit, cfg.RateWindow)(
			middleware.Metrics()(
				middleware.JSONLogging(logger)(
					mux,
				),
			),
		),
	)

	srv := &http.Server{
		Addr:         cfg.Addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "addr", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server error", "error", err)
			os.Exit(1)
		}
	}()

	_ = openBrowser(cfg.BaseURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Graceful shutdown error", "error", err)
	}
	logger.Info("Server stopped")
}

func openBrowser(url string) error {
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		if os.Getenv("BROWSER") != "" {
			parts := strings.Split(os.Getenv("BROWSER"), " ")
			return exec.Command(parts[0], append(parts[1:], url)...).Start()
		}
		return exec.Command("xdg-open", url).Start()
	}
}
