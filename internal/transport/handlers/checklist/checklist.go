package checklist

import (
	"api-gateway/internal/transport/middleware"
	api "api-gateway/pkg/api/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type HandlerChecklist struct {
	eventClient api.EventServiceClient
}

func NewHandlerChecklist(eventClient api.EventServiceClient) *HandlerChecklist {
	return &HandlerChecklist{
		eventClient: eventClient,
	}
}

// GetEventChecklist godoc
// @Summary Получить чеклист мероприятия
// @Tags interaction
// @Accept json
// @Produce json
// @Param event_id path string true "ID мероприятия"
// @Security BearerAuth
// @Success 200 {object} api.GetEventChecklistResponse "Возвращает список предметов в чеклисте"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/checklist [get]
func (h *HandlerChecklist) GetEventChecklist(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("checklist.GetEventChecklist user not found in context")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("checklist.GetEventChecklist empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	grpcReq := api.GetEventChecklistRequest{
		EventId: eventID,
	}
	resp, err := h.eventClient.GetEventChecklist(ctx, &grpcReq)
	if err != nil {
		l.Error("checklist.GetEventChecklist error",
			"err", err.Error())
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

type AddChecklistDTO struct {
	Title    string `json:"title"`
	Quantity int32  `json:"quantity"`
	Unit     string `json:"unit"`
}

// AddChecklistItem godoc
// @Summary Добавить предмет в чеклист ивента
// @Tags interaction
// @Accept json
// @Produce json
// @Param event_id path string true "ID мероприятия"
// @Param request body AddChecklistDTO true "Данные предмета"
// @Security BearerAuth
// @Success 200 {object} api.AddChecklistItemResponse "Возвращает id созданного товара в чеклисте"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/checklist [post]
func (h *HandlerChecklist) AddChecklistItem(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("checklist.AddChecklistItem user not found in context")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)
	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("checklist.AddChecklistItem empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}

	var reqDTO AddChecklistDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("checklist.AddChecklistItem error",
			"err", err.Error())
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	grpcReq := api.AddChecklistItemRequest{
		EventId:  eventID,
		Quantity: reqDTO.Quantity,
		Unit:     reqDTO.Unit,
		Title:    reqDTO.Title,
	}
	resp, err := h.eventClient.AddChecklistItem(ctx, &grpcReq)
	if err != nil {
		l.Error("checklist.AddChecklistItem error",
			"err", err.Error())
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// RemoveChecklistItem godoc
// @Summary Удалить предмет из чеклиста
// @Tags interaction
// @Accept json
// @Produce json
// @Param event_id path string true "ID мероприятия"
// @Param item_id path string true "ID предмета"
// @Security BearerAuth
// @Success 200 {object} api.RemoveChecklistItemResponse "Возвращает id созданного товара в чеклисте"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/checklist/{item_id} [delete]
func (h *HandlerChecklist) RemoveChecklistItem(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("checklist.RemoveChecklistItem user not found in context")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("checklist.RemoveChecklistItem empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}
	itemID := r.PathValue("item_id")
	if itemID == "" {
		l.Error("checklist.RemoveChecklistItem empty item_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing item_id")
		return
	}
	grpcReq := api.RemoveChecklistItemRequest{
		EventId: eventID,
		ItemId:  itemID,
	}
	resp, err := h.eventClient.RemoveChecklistItem(ctx, &grpcReq)
	if err != nil {
		l.Error("checklist.RemoveChecklistItem error",
			"err", err.Error())
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

type MarkItemDTO struct {
	BuyerID     *string `json:"buyer_id"`
	IsPurchased *bool   `json:"is_purchased"`
}

// MarkItemPurchased godoc
// @Summary Пометить предмет как купленный в чеклисте
// @Description Также можно добавить id ответственного за покупку (buyer_id), а также передать стату куплено/не куплно
// @Tags interaction
// @Accept json
// @Produce json
// @Param event_id path string true "ID мероприятия"
// @Param item_id path string true "ID предмета"
// @Param request body MarkItemDTO true "Данные покупки"
// @Security BearerAuth
// @Success 200 {object} api.MarkItemPurchasedResponse "Возвращает статус success"
// @Failure 400,401,500 {object} response.ErrorResponse
// @Router /v1/events/{event_id}/checklist/{item_id}/purchase [patch]
func (h *HandlerChecklist) MarkItemPurchased(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())
	userUUID, ok := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
	if !ok {
		l.Error("checklist.MarkItemPurchased user not found in context")
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "unauthorized")
		return
	}
	userID := userUUID.String()
	ctx := metadata.AppendToOutgoingContext(r.Context(), "x-user-id", userID)

	eventID := r.PathValue("event_id")
	if eventID == "" {
		l.Error("checklist.MarkItemPurchased empty event_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing event_id")
		return
	}
	itemID := r.PathValue("item_id")
	if itemID == "" {
		l.Error("checklist.MarkItemPurchased empty item_id")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "missing item_id")
		return
	}

	var reqDTO MarkItemDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("checklist.MarkItemPurchased error",
			"err", err.Error())
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}
	grpcReq := api.MarkItemPurchasedRequest{
		EventId:     eventID,
		ItemId:      itemID,
		IsPurchased: reqDTO.IsPurchased,
		BuyerId:     reqDTO.BuyerID,
	}
	resp, err := h.eventClient.MarkItemPurchased(ctx, &grpcReq)
	if err != nil {
		l.Error("checklist.MarkItemPurchased error",
			"err", err.Error())
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}
	response.JSON(w, http.StatusOK, resp)
}
