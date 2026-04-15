package interaction

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

type HandlerInteraction struct {
	eventClient api.EventServiceClient
}

func NewHandlerInteraction(eventClient api.EventServiceClient) *HandlerInteraction {
	return &HandlerInteraction{
		eventClient: eventClient,
	}
}

type JoinEventDTO struct {
	EventCode string `json:"event_code"`
}

// JoinEvent godoc
// @Summary Присоединиться к ивенту
// @Description в теле запроса отправляется eventCode
// @Tags interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.JoinEventResponse "Возвращает id ивента и статус успешно/безуспешно"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/join [post]
func (h *HandlerInteraction) JoinEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("interaction.JoinEvent user id is required")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	var reqDTO JoinEventDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("Interaction.JoinEvent bad json",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}

	grpcReq := api.JoinEventRequest{
		EventCode: reqDTO.EventCode,
	}
	resp, err := h.eventClient.JoinEvent(ctx, &grpcReq)
	if err != nil {
		l.Error("interaction.JoinEvent error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "INTERNAL ERROR")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// LeaveEvent godoc
// @Summary Покинуть ивент
// @Tags interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.LeaveEventResponse "Возвращает статус (Успешно/не успешно)"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/leave [post]
func (h *HandlerInteraction) LeaveEvent(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("interaction.LeaveEvent user id is required")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)
	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("interaction.LeaveEvent empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}
	grpcReq := api.LeaveEventRequest{
		EventId: eventID,
	}
	resp, err := h.eventClient.LeaveEvent(ctx, &grpcReq)
	if err != nil {
		l.Error("interaction.LeaveEvent error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "INTERNAL_ERROR")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

type CreateLinkInviteDTO struct {
	InviteType string    `json:"invite_type"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// CreateInviteLink godoc
// @Summary Создать ссылку приглашение
// @Description Тип приглашения (single/multi/unlimited)
// @Tags interaction
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.CreateInviteLinkResponse "Возвращает event_code"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/invites [post]
func (h *HandlerInteraction) CreateInviteLink(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("interaction.CreateInviteLink user id is required")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "user id is not found")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)
	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("interaction.CreateInviteLink empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	var reqDTO CreateLinkInviteDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("Interaction.CreateInviteLink bad json",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "bad json")
		return
	}
	grpcReq := api.CreateInviteLinkRequest{
		EventId:    eventID,
		InviteType: reqDTO.InviteType,
		ExpiresAt:  timestamppb.New(reqDTO.ExpiresAt),
	}

	resp, err := h.eventClient.CreateInviteLink(ctx, &grpcReq)
	if err != nil {
		l.Error("interaction.CreateInviteLink error",
			slog.String("error", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "INTERNAL_ERROR")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
