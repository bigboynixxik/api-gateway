package event

import (
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"net/http"
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
	logs := logger.FromContext(r.Context())

	resp, err := h.eventClient.CreateEvent(r.Context(), &api.CreateEventRequest{})
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
}
