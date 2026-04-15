package event

import (
	"api-gateway/internal/transport/middleware"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type HandlerEvent struct {
	eventClient api.EventServiceClient
}

func NewHandlerEvent(eventClient api.EventServiceClient) *HandlerEvent {
	return &HandlerEvent{
		eventClient: eventClient,
	}
}

type CreateEventDTO struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	IsPrivate       bool      `json:"is_private"`
	DurationMinutes int32     `json:"duration_minutes"`
	StartsAt        time.Time `json:"starts_at"`
	LocationName    string    `json:"location_name"`
	MaxParticipants int32     `json:"max_participants"`
	LocationCoords  *string   `json:"location_coords"`
}

// CreateEvent godoc
// @Summary Создать новое мероприятие
// @Description Создает ивент. Обязательно передавать Bearer токен в заголовке Authorization.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateEventDTO true "Данные мероприятия"
// @Success 200 {object} response.JSON "Возвращает ID созданного ивента"
// @Failure 400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} response.ErrorResponse "Не авторизован"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /v1/event/create [post]
func (h *HandlerEvent) CreateEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("event.CreateEvent user id is required")
		response.Error(w, http.StatusUnauthorized, "UNATHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	var reqDTO CreateEventDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("event.CreateEvent bad json",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}

	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	parsedTime, err := time.Parse(time.RFC3339, reqDTO.StartsAt.Format(time.RFC3339))
	if err != nil {
		l.Error("event.CreateEvent bad start time",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "bad start time")
		return
	}

	grpcReq := api.CreateEventRequest{
		Title:           reqDTO.Title,
		Description:     reqDTO.Description,
		IsPrivate:       reqDTO.IsPrivate,
		DurationMinutes: reqDTO.DurationMinutes,
		StartsAt:        timestamppb.New(parsedTime),
		LocationName:    reqDTO.LocationName,
		MaxParticipants: reqDTO.MaxParticipants,
		LocationCoords:  reqDTO.LocationCoords,
	}

	resp, err := h.eventClient.CreateEvent(ctx, &grpcReq)
	if err != nil {
		l.Error("event.CreateEvent internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *HandlerEvent) ListUserEvents(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("event.ListUserEvents user id is required")
		response.Error(w, http.StatusUnauthorized, "UNATHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)
	resp, err := h.eventClient.ListUserEvents(ctx, &api.ListUserEventsRequest{})
	if err != nil {
		l.Error("event.ListUserEvents internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *HandlerEvent) GetEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

}
