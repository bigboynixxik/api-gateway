package app

import (
	"api-gateway/internal/transport/handlers/checklist"
	"api-gateway/internal/transport/handlers/interaction"
	"api-gateway/internal/transport/handlers/participants"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/transport/handlers"
	"api-gateway/internal/transport/handlers/authorization"
	"api-gateway/internal/transport/handlers/event"
	"api-gateway/internal/transport/interceptor"
	"api-gateway/internal/transport/middleware"
	auth "api-gateway/pkg/api/auth/v1"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/closer"
	"api-gateway/pkg/config"
	"api-gateway/pkg/logger"

	googleGrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	eventClient      api.EventServiceClient
	authClient       auth.AuthServiceClient
	httpServer       *http.Server
	eventServiceAddr string
	authServiceAddr  string
	httpPort         string
	logs             *slog.Logger
	closer           *closer.Closer
}

func NewApp(_ context.Context) (*App, error) {
	cfg, err := config.LoadConfig(".env")
	if err != nil {
		return nil, fmt.Errorf("app.New failed to load config: %w", err)
	}

	logger.Setup(cfg.AppEnv)
	logs := logger.With("service", "api-gateway")
	logs.Info("initializing layers",
		"env", cfg.AppEnv,
		"HTTPPort", cfg.HTTPPort,
		"Event service addr", cfg.EventServiceAddr,
		"Auth service addr", cfg.AuthServiceAddr)

	eventConn, err := googleGrpc.NewClient(cfg.EventServiceAddr,
		googleGrpc.WithTransportCredentials(insecure.NewCredentials()),
		googleGrpc.WithUnaryInterceptor(interceptor.LoggingClientInterceptor(logs)))
	if err != nil {
		return nil, fmt.Errorf("app.New unable to connect to gRPC server: %w", err)
	}

	eventClient := api.NewEventServiceClient(eventConn)

	authConn, err := googleGrpc.NewClient(cfg.AuthServiceAddr,
		googleGrpc.WithTransportCredentials(insecure.NewCredentials()),
		googleGrpc.WithUnaryInterceptor(interceptor.LoggingClientInterceptor(logs)))
	if err != nil {
		return nil, fmt.Errorf("app.New unable to connect to gRPC server: %w", err)
	}
	authClient := auth.NewAuthServiceClient(authConn)

	authMW := middleware.NewAuthMiddleware(authClient, cfg.JWTSecret)

	eventHandler := event.NewHandlerEvent(eventClient)
	authHandler := authorization.NewAuthHandler(authClient)
	interactionHandler := interaction.NewHandlerInteraction(eventClient)
	participantsHandler := participants.NewHandlerParticipant(eventClient, authClient)
	checklistHandler := checklist.NewHandlerChecklist(eventClient)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /_info", handlers.InfoHandler)

	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("GET /users/{login}", authHandler.GetUserInfoByLogin)
	mux.HandleFunc("POST /users/info", authHandler.GetUsersInfo)

	mux.HandleFunc("GET /v1/events/my", authMW.AuthMiddleware(eventHandler.ListUserEvents))
	mux.HandleFunc("POST /v1/events", authMW.AuthMiddleware(eventHandler.CreateEvent))
	mux.HandleFunc("PATCH /v1/events/{event_id}", authMW.AuthMiddleware(eventHandler.UpdateEvent))
	mux.HandleFunc("DELELTE /v1/events/{event_id}", authMW.AuthMiddleware(eventHandler.CancelEvent))
	mux.HandleFunc("GET /v1/events/{event_id}", eventHandler.GetEvent)
	mux.HandleFunc("GET /v1/events", eventHandler.ListEvents)

	mux.HandleFunc("POST /v1/events/{event_id}/invites", authMW.AuthMiddleware(interactionHandler.CreateInviteLink))
	mux.HandleFunc("POST /v1/events/{event_id}/leave", authMW.AuthMiddleware(interactionHandler.LeaveEvent))
	mux.HandleFunc("POST /v1/events/join", authMW.AuthMiddleware(interactionHandler.JoinEvent))

	mux.HandleFunc("GET /v1/events/{event_id}/participants", participantsHandler.GetEventParticipant)
	mux.HandleFunc("DELETE /v1/events/{event_id}/participants/{participant_id}", authMW.AuthMiddleware(participantsHandler.RemoveParticipant))

	mux.HandleFunc("GET /v1/events/{event_id}/checklist", authMW.AuthMiddleware(checklistHandler.GetEventChecklist))
	mux.HandleFunc("POST /v1/events/{event_id}/checklist", authMW.AuthMiddleware(checklistHandler.AddChecklistItem))
	mux.HandleFunc("DELETE /v1/events/{event_id}/checklist/{item_id}", authMW.AuthMiddleware(checklistHandler.RemoveChecklistItem))
	mux.HandleFunc("POST /v1/events/{event_id}/checklist/{item_id}/purchase", authMW.AuthMiddleware(checklistHandler.MarkItemPurchased))

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: mux,
	}

	cl := closer.New()

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing event service connection")
		return eventConn.Close()
	})

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing auth service connection")
		return authConn.Close()
	})

	cl.Add(func(ctx context.Context) error {
		slog.Info("closing http server")
		return httpServer.Shutdown(ctx)
	})

	return &App{
		eventClient:      eventClient,
		authClient:       authClient,
		httpServer:       httpServer,
		httpPort:         cfg.HTTPPort,
		eventServiceAddr: cfg.EventServiceAddr,
		authServiceAddr:  cfg.AuthServiceAddr,
		logs:             logs,
		closer:           cl,
	}, nil
}

func (a *App) Run() {
	errCh := make(chan error)

	go func() {
		a.logs.Info("starting http server")
		if err := a.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	a.logs.Info("App.Run starting server", "port", a.httpPort)

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		a.logs.Error("app.run server startup failed", "error", err)
	case sig := <-quit:
		a.logs.Info("app.run server shutdown", "signal", sig)
	}

	a.logs.Info("shutting down servers")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := a.closer.Close(shutdownCtx); err != nil {
		a.logs.Error("close server shutdown failed", "error", err)
	}

	fmt.Println("Server stopped")
}
