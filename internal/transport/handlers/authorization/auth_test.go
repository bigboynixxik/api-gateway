package authorization

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	auth "api-gateway/pkg/api/auth/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ручной мок для gRPC
type mockAuthClient struct {
	auth.AuthServiceClient
	loginFunc              func(ctx context.Context, in *auth.LoginRequest, opts ...grpc.CallOption) (*auth.LoginResponse, error)
	registerFunc           func(ctx context.Context, in *auth.RegisterRequest, opts ...grpc.CallOption) (*auth.RegisterResponse, error)
	getUserInfoByLoginFunc func(ctx context.Context, in *auth.GetUserInfoByLoginRequest, opts ...grpc.CallOption) (*auth.UserInfo, error)
}

func (m *mockAuthClient) Login(ctx context.Context, in *auth.LoginRequest, opts ...grpc.CallOption) (*auth.LoginResponse, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "mock not implemented")
}

func (m *mockAuthClient) Register(ctx context.Context, in *auth.RegisterRequest, opts ...grpc.CallOption) (*auth.RegisterResponse, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "mock not implemented")
}

func (m *mockAuthClient) GetUserInfoByLogin(ctx context.Context, in *auth.GetUserInfoByLoginRequest, opts ...grpc.CallOption) (*auth.UserInfo, error) {
	if m.getUserInfoByLoginFunc != nil {
		return m.getUserInfoByLoginFunc(ctx, in, opts...)
	}
	return nil, status.Error(codes.Unimplemented, "mock not implemented")
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        interface{}
		mockResp       *auth.LoginResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Успешный логин",
			reqBody:        LoginDTO{Email: "test@mail.com", Password: "123"},
			mockResp:       &auth.LoginResponse{AccessToken: "token123"},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Юзер не найден",
			reqBody:        LoginDTO{Email: "notfound@mail.com", Password: "123"},
			mockResp:       nil,
			mockErr:        status.Error(codes.NotFound, "not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{
				loginFunc: func(ctx context.Context, in *auth.LoginRequest, opts ...grpc.CallOption) (*auth.LoginResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			handler := NewAuthHandler(client)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		reqBody        interface{}
		mockResp       *auth.RegisterResponse
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Успешная регистрация",
			reqBody:        RegisterDTO{Email: "new@mail.com", Login: "newbie", Name: "Ivan", Password: "123"},
			mockResp:       &auth.RegisterResponse{AccessToken: "token_reg"},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Кривой JSON",
			reqBody:        "invalid",
			mockResp:       nil,
			mockErr:        nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Уже существует",
			reqBody:        RegisterDTO{Email: "exist@mail.com", Login: "exist", Name: "Ivan", Password: "123"},
			mockResp:       nil,
			mockErr:        status.Error(codes.AlreadyExists, "already exists"),
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{
				registerFunc: func(ctx context.Context, in *auth.RegisterRequest, opts ...grpc.CallOption) (*auth.RegisterResponse, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			handler := NewAuthHandler(client)

			var body []byte
			if str, ok := tt.reqBody.(string); ok {
				body = []byte(str)
			} else {
				body, _ = json.Marshal(tt.reqBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.Register(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthHandler_GetUserInfoByLogin(t *testing.T) {
	tests := []struct {
		name           string
		login          string
		mockResp       *auth.UserInfo
		mockErr        error
		expectedStatus int
	}{
		{
			name:           "Успешный поиск",
			login:          "vasya",
			mockResp:       &auth.UserInfo{Id: "1", Login: "vasya", Name: "Vasiliy"},
			mockErr:        nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Не найден",
			login:          "ghost",
			mockResp:       nil,
			mockErr:        status.Error(codes.NotFound, "not found"),
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAuthClient{
				getUserInfoByLoginFunc: func(ctx context.Context, in *auth.GetUserInfoByLoginRequest, opts ...grpc.CallOption) (*auth.UserInfo, error) {
					return tt.mockResp, tt.mockErr
				},
			}
			handler := NewAuthHandler(client)

			req := httptest.NewRequest(http.MethodGet, "/v1/users/"+tt.login, nil)
			if tt.login != "" {
				req.SetPathValue("login", tt.login)
			}

			w := httptest.NewRecorder()
			handler.GetUserInfoByLogin(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
