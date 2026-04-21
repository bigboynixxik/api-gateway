package interaction

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-gateway/internal/transport/middleware"
	api "api-gateway/pkg/api/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
)

// Ручной мок для event-клиента
type mockEventClient struct {
	api.EventServiceClient
	joinEventFunc func(ctx context.Context, in *api.JoinEventRequest, opts ...grpc.CallOption) (*api.JoinEventResponse, error)
}

func (m *mockEventClient) JoinEvent(ctx context.Context, in *api.JoinEventRequest, opts ...grpc.CallOption) (*api.JoinEventResponse, error) {
	if m.joinEventFunc != nil {
		return m.joinEventFunc(ctx, in, opts...)
	}
	return nil, nil
}

func TestHandlerInteraction_JoinEvent(t *testing.T) {
	validUUID := uuid.New()

	tests := []struct {
		name           string
		injectUser     bool
		reqBody        interface{}
		mockResp       *api.JoinEventResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Успешный джоин",
			injectUser:     true,
			reqBody:        JoinEventDTO{EventCode: "SECRET123"},
			mockResp:       &api.JoinEventResponse{EventId: "event-1", Success: true},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Нет юзера в контексте (ошибка 401)",
			injectUser:     false, // специально не кладем uuid в контекст
			reqBody:        JoinEventDTO{EventCode: "SECRET123"},
			mockResp:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Кривой JSON",
			injectUser:     true,
			reqBody:        "это-не-джсон",
			mockResp:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Ошибка со стороны gRPC сервера",
			injectUser:     true,
			reqBody:        JoinEventDTO{EventCode: "SECRET123"},
			mockResp:       nil,
			mockErr:        errors.New("internal grpc error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockEventClient{
				joinEventFunc: func(ctx context.Context, in *api.JoinEventRequest, opts ...grpc.CallOption) (*api.JoinEventResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			handler := NewHandlerInteraction(client)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/events/join", bytes.NewReader(body))

			if tt.injectUser {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, validUUID)
				req = req.WithContext(ctx)
			}

			w := httptest.NewRecorder()

			handler.JoinEvent(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
