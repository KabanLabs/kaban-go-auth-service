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
	status      int
	body        *bytes.Buffer
	wroteHeader bool
	captured    bool
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.status = code
		rw.wroteHeader = true
	}
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}

	rw.body.Write(b)

	return len(b), nil
}

func cookieMiddleware(next http.Handler, refreshTokenTTL time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login" || r.URL.Path == "/refresh" {
			rw := &responseWriterWrapper{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				status:         http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			if rw.status >= 200 && rw.status < 300 {
				var resp map[string]interface{}

				bodyBytes := rw.body.Bytes()
				if err := json.Unmarshal(bodyBytes, &resp); err == nil {
					if refreshToken, exists := resp["refreshToken"].(string); exists && refreshToken != "" {
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
						if rw.wroteHeader {
							w.WriteHeader(rw.status)
						}
						json.NewEncoder(w).Encode(resp)
						return
					}
				}
			}

			if rw.wroteHeader {
				w.WriteHeader(rw.status)
			}

			_, err := w.Write(rw.body.Bytes())

			if err != nil {
				return
			}

		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func New(
	ctx context.Context,
	log *slog.Logger,
	httpPort int,
	grpcPort int,
	refreshTokenTTL time.Duration,
) *App {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := ssov1.RegisterAuthHandlerFromEndpoint(ctx, mux, fmt.Sprintf("localhost:%d", grpcPort), opts)

	if err != nil {
		panic(err)
	}

	cookieMiddleware := cookieMiddleware(mux, refreshTokenTTL)

	return &App{
		server:  mux,
		handler: cookieMiddleware,
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
