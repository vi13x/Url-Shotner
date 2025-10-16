package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"URL_shortener/internal/httpapi"
	"URL_shortener/internal/shortener"
	"URL_shortener/internal/storage"
)

func main() {
	addr := getEnv("ADDR", ":8080")
	baseURL := getEnv("BASE_URL", "http://localhost"+addr)

	store := storage.NewMemoryStorage()
	svc := shortener.NewService(store)
	api := httpapi.NewHandler(svc, baseURL)

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("web")))
	mux.Handle("/api/", api.Routes())
	mux.Handle("/s/", api.Routes())

	srv := &http.Server{Addr: addr, Handler: loggingMiddleware(mux)}

	go func() {
		log.Printf("Listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

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

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
