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
	body   bytes.Buffer
	status int
}

func (rw *responseWriterWrapper) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriterWrapper) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
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

	cookieMiddleware := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w}

		mux.ServeHTTP(rw, r)

		if (r.URL.Path == "/login" || r.URL.Path == "/refresh") && rw.status >= 200 && rw.status < 300 {
			var resp struct {
				RefreshToken string `json:"refresh_token"`
			}

			if err := json.Unmarshal(rw.body.Bytes(), &resp); err == nil && resp.RefreshToken != "" {
				http.SetCookie(rw, &http.Cookie{
					Name:     "refresh_token",
					Value:    resp.RefreshToken,
					HttpOnly: true,
					Path:     "/",
					MaxAge:   int(refreshTokenTTL.Seconds()),
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
				})

				resp.RefreshToken = ""
				rw.body.Reset()
				_ = json.NewEncoder(rw).Encode(resp)
			}
		}
	})

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
