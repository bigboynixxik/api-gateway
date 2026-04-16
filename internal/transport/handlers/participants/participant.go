package participants

import (
	"api-gateway/internal/transport/middleware"
	auth "api-gateway/pkg/api/auth/v1"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type HandlerParticipant struct {
	eventClient api.EventServiceClient
	authClient  auth.AuthServiceClient
}

func NewHandlerParticipant(eventClient api.EventServiceClient, authClient auth.AuthServiceClient) *HandlerParticipant {
	return &HandlerParticipant{
		eventClient: eventClient,
		authClient:  authClient,
	}
}

type ParticipantResponseDTO struct {
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Login  string `json:"login"`
	Role   string `json:"role"`
	Status string `json:"status"`
}

// GetEventParticipant godoc
// @Summary Получить список участников ивента
// @Tags participants
// @Accept json
// @Produce json
// @Param event_id path string true "ID мероприятия"
// @Success 200 {array} ParticipantResponseDTO "Возвращает список участников с именами"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/participants [get]
func (h *HandlerParticipant) GetEventParticipant(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("participants.GetEventParticipant empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	eventResp, err := h.eventClient.GetEventParticipants(r.Context(), &api.GetEventParticipantsRequest{EventId: eventID})
	if err != nil {
		l.Error("participants.GetEventParticipant err",
			slog.String("err", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	var userIDs []string
	for _, p := range eventResp.Participants {
		userIDs = append(userIDs, p.ParticipantId)
	}

	if len(userIDs) == 0 {
		l.Warn("participants.GetEventParticipant empty userIDs")
		response.JSON(w, http.StatusOK, []ParticipantResponseDTO{})
		return
	}

	authResp, err := h.authClient.GetUsersInfo(r.Context(), &auth.GetUsersInfoRequest{UserIds: userIDs})
	if err != nil {
		l.Error("participant.GetEventParticipant failed to get users info", slog.String("err", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to fetch user details")
		return
	}

	var result []ParticipantResponseDTO
	for _, p := range eventResp.Participants {
		userInfo, ok := authResp.Users[p.ParticipantId]

		name := "No name"
		login := ""
		if ok {
			name = userInfo.Name
			login = userInfo.Login
		}

		result = append(result, ParticipantResponseDTO{
			UserID: p.ParticipantId,
			Name:   name,
			Login:  login,
			Role:   p.Role,
			Status: p.Status,
		})
	}
	response.JSON(w, http.StatusOK, result)
}

// RemoveParticipant godoc
// @Summary удалить участника из мероприятия
// @Tags participants
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} api.RemoveParticipantResponse "Возвращает статус succes"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/participants/{participant_id} [delete]
func (h *HandlerParticipant) RemoveParticipant(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("participant.RemoveParticipant user not found in context")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("participants.RemoveParticipant empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	participantID := r.PathValue("participant_id")
	if participantID == "" {
		l.Error("participants.RemoveParticipant empty participant_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing participant_id")
		return
	}

	grpcReq := api.RemoveParticipantRequest{
		EventId:       eventID,
		ParticipantId: participantID,
	}
	resp, err := h.eventClient.RemoveParticipant(ctx, &grpcReq)
	if err != nil {
		l.Error("participant.RemoveParticipant error",
			slog.String("err", err.Error()))
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
