package event

import (
	"api-gateway/internal/transport/middleware"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type HandlerEvent struct {
	eventClient api.EventServiceClient
}

func NewHandlerEvent(eventClient api.EventServiceClient) *HandlerEvent {
	return &HandlerEvent{
		eventClient: eventClient,
	}
}

func (h *HandlerEvent) CreateEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("event.CreateEvent user id is required")
		response.Error(w, http.StatusUnauthorized, "UNATHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	var req api.CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		l.Error("event.CreateEvent bad json",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}

	md := metadata.Pairs("x-user-id", userID)
	ctx := metadata.NewOutgoingContext(r.Context(), md)

	resp, err := h.eventClient.CreateEvent(ctx, &req)
	if err != nil {
		l.Error("event.CreateEvent internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
