package event

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"api-gateway/internal/transport/middleware"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"

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
// @Success 200 {object} api.CreateEventResponse "Возвращает ID созданного ивента"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events [post]
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

	grpcReq := api.CreateEventRequest{
		Title:           reqDTO.Title,
		Description:     reqDTO.Description,
		IsPrivate:       reqDTO.IsPrivate,
		DurationMinutes: reqDTO.DurationMinutes,
		StartsAt:        timestamppb.New(reqDTO.StartsAt),
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

// ListUserEvents godoc
// @Summary Список мероприятий пользователя
// @Desctiption получить список мероприятий конкретного пользователя. Обязательно передавать Bearer токен в заголовке Authorization.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.ListUserEventsResponse "Возвращает список мероприятий пользователя"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/my [get]
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

// GetEvent godoc
// @Summary получить информацию об ивенте
// @Description отправляется id ивента, возвращается информация о нём
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} api.GetEventResponse "Возвращает список: информация об ивенте"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id} [get]
func (h *HandlerEvent) GetEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("event.GetEvent empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	resp, err := h.eventClient.GetEvent(r.Context(), &api.GetEventRequest{
		Id: eventID,
	})
	if err != nil {
		l.Error("event.GetEvent internal error", slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

type UpdateEventDTO struct {
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	StartsAt       *time.Time `json:"starts_at"`
	LocationName   *string    `json:"location_name"`
	LocationCoords *string    `json:"location_coords"`
}

// UpdateEvent godoc
// @Summary Обновить мероприятие
// @Description Частичное обновление данных (PATCH). Обязательно Bearer токен.
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param event_id path string true "ID мероприятия"
// @Param request body UpdateEventDTO true "Новые данные"
// @Success 200 {object} api.UpdateEventResponse
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id} [patch]
func (h *HandlerEvent) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("event.ListUserEvents user id is required")
		response.Error(w, http.StatusUnauthorized, "UNATHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("event.GetEvent empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}
	var reqDTO UpdateEventDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("event.UpdateEvent bad json",
			slog.String("error", err.Error()))
	}

	grpcReq := api.UpdateEventRequest{
		EventId:        eventID,
		Title:          reqDTO.Title,
		Description:    reqDTO.Description,
		LocationName:   reqDTO.LocationName,
		LocationCoords: reqDTO.LocationCoords,
	}
	if reqDTO.StartsAt != nil {
		grpcReq.StartsAt = timestamppb.New(*reqDTO.StartsAt)
	}

	resp, err := h.eventClient.UpdateEvent(ctx, &grpcReq)
	if err != nil {
		l.Error("event.UpdateEvent internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)

}

// CancelEvent godoc
// @Summary Отменить ивент
// @Tags events
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.CancelEventResponse "Возвращает список: обновлённые данные об ивенте"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id} [delete]
func (h *HandlerEvent) CancelEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("event.ListUserEvents user id is required")
		response.Error(w, http.StatusUnauthorized, "UNATHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("event.GetEvent empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}
	grpcReq := api.CancelEventRequest{
		EventId: eventID,
	}
	resp, err := h.eventClient.CancelEvent(ctx, &grpcReq)
	if err != nil {
		l.Error("event.CancelEvent internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

type ListEventsDTO struct {
	Title        *string    `json:"title"`
	Description  *string    `json:"description"`
	StartAfter   *time.Time `json:"start_after"`
	StartsBefore *time.Time `json:"starts_before"`
	LocationName *string    `json:"location_name"`
}

// ListEvents godoc
// @Summary Получить список ивентов
// @Description Получение списка ивентов. Также можно через query-параметры указать: поиск по названию (title), поиск по описанию (description); фильтрация по дате: (starts_after, starts_before), фильтр по локации (location_name)
// @Tags events
// @Accept json
// @Produce json
// @Success 200 {object} api.ListEventsResponse "Возвращает список ивентов"
// @Failure 400,500 {object} response.ErrorResponse
// @Router /v1/events [get]
func (h *HandlerEvent) ListEvents(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	var reqDTO ListEventsDTO
	title := r.URL.Query().Get("title")
	if title != "" {
		reqDTO.Title = &title
	}
	description := r.URL.Query().Get("description")
	if description != "" {
		reqDTO.Description = &description
	}
	startsAfter := r.URL.Query().Get("start_after")
	if startsAfter != "" {
		startAfterTime, err := time.Parse(time.RFC3339, startsAfter)
		if err != nil {
			l.Error("event.ListEvents start_after error",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusBadRequest, "INTERNAL_ERROR", "internal error")
			return
		}
		reqDTO.StartAfter = &startAfterTime
	}
	startsBefore := r.URL.Query().Get("starts_before")
	if startsBefore != "" {
		startsBeforeTime, err := time.Parse(time.RFC3339, startsBefore)
		if err != nil {
			l.Error("event.ListEvents end_after error",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			return
		}
		reqDTO.StartsBefore = &startsBeforeTime
	}
	locationName := r.URL.Query().Get("location_name")
	if locationName != "" {
		reqDTO.LocationName = &locationName
	}

	grpcReq := api.ListEventsRequest{
		Title:        reqDTO.Title,
		Description:  reqDTO.Description,
		LocationName: reqDTO.LocationName,
	}

	if reqDTO.StartAfter != nil {
		grpcReq.StartsAfter = timestamppb.New(*reqDTO.StartAfter)
	}
	if reqDTO.StartsBefore != nil {
		grpcReq.StartsBefore = timestamppb.New(*reqDTO.StartsBefore)
	}

	resp, err := h.eventClient.ListEvents(r.Context(), &grpcReq)
	if err != nil {
		l.Error("event.ListEvents internal error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
