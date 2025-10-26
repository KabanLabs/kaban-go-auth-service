package gateway

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	ssov1 "github.com/VACdotCS/protos/gen/go/sso"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	server  *runtime.ServeMux
	handler http.Handler
	port    int
	log     *slog.Logger
}

type responseWriterWrapper struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
	header http.Header
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	return &responseWriterWrapper{
		ResponseWriter: w,
		status:         0,
		body:           &bytes.Buffer{},
		header:         make(http.Header),
	}
}

func (rw *responseWriterWrapper) Header() http.Header {
	return rw.header
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if rw.status == 0 {
		rw.status = code
	}
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	return rw.body.Write(b)
}

func cookieMiddleware(next http.Handler, refreshTokenTTL time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" || r.URL.Path == "/refresh" {
			rw := newResponseWriterWrapper(w)

			next.ServeHTTP(rw, r)

			if rw.status == 0 {
				rw.status = http.StatusOK
			}

			for k, v := range rw.header {
				for _, vv := range v {
					w.Header().Add(k, vv)
				}
			}

			if rw.status >= 200 && rw.status < 300 {
				var resp map[string]interface{}
				bodyBytes := rw.body.Bytes()

				if err := json.Unmarshal(bodyBytes, &resp); err == nil {
					if refreshToken, ok := resp["refreshToken"].(string); ok && refreshToken != "" {
						http.SetCookie(w, &http.Cookie{
							Name:     "refreshToken",
							Value:    refreshToken,
							HttpOnly: true,
							Path:     "/",
							MaxAge:   int(refreshTokenTTL.Seconds()),
							Secure:   true,
							SameSite: http.SameSiteLaxMode,
						})

						delete(resp, "refreshToken")

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(rw.status)
						_ = json.NewEncoder(w).Encode(resp)
						return
					}
				}
			}

			w.WriteHeader(rw.status)
			_, _ = w.Write(rw.body.Bytes())
			return
		}

		// остальные запросы без обёртки
		next.ServeHTTP(w, r)
	})
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Access-Control-Allow-Origin header to allow requests from any origin
		// For production, replace "*" with specific origins like "http://example.com"
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests (OPTIONS method)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func New(
	ctx context.Context,
	log *slog.Logger,
	httpPort int,
	grpcPort int,
	corsEnabled bool,
	refreshTokenTTL time.Duration,
) *App {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := ssov1.RegisterAuthHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)

	if err != nil {
		panic(err)
	}

	var httpHandler http.Handler

	cookieMiddleware := cookieMiddleware(mux, refreshTokenTTL)

	if corsEnabled {
		httpHandler = enableCORS(cookieMiddleware)
	} else {
		httpHandler = cookieMiddleware
	}

	return &App{
		server:  mux,
		handler: httpHandler,
		port:    httpPort,
		log:     log,
	}
}

func (app *App) Run() {
	logger := app.log.With(
		slog.String("op", "httpApp.Run"),
		slog.Int("port", app.port),
	)

	logger.Info("Http gateway is running")

	if err := http.ListenAndServe(fmt.Sprintf(":%d", app.port), app.handler); err != nil {
		panic(err)
	}
}
