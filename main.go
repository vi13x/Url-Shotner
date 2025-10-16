package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"URL_shortener/internal/httpapi"
	"URL_shortener/internal/shortener"
	"URL_shortener/internal/storage"
)

//go:embed web/*
var embeddedWeb embed.FS

func main() {
	addr := ":8080"
	baseURL := "http://localhost" + addr

	store := storage.NewMemoryStorage()
	svc := shortener.NewService(store)
	api := httpapi.NewHandler(svc, baseURL)

	mux := http.NewServeMux()
	webFS, err := fs.Sub(embeddedWeb, "web")
	if err != nil {
		log.Fatalf("embed sub fs error: %v", err)
	}
	staticFS, err := fs.Sub(webFS, "static")
	if err != nil {
		log.Fatalf("embed static fs error: %v", err)
	}
	fsys := http.FS(webFS)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
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

	srv := &http.Server{Addr: addr, Handler: loggingMiddleware(mux)}

	go func() {
		log.Printf("Listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	_ = openBrowser(baseURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
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
