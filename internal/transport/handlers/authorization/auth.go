package authorization

import (
	auth "api-gateway/pkg/api/auth/v1"
	"api-gateway/pkg/logger"
	"api-gateway/pkg/response"
	"encoding/json"
	"log/slog"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthHandler struct {
	authClient auth.AuthServiceClient
}

func NewAuthHandler(authClient auth.AuthServiceClient) *AuthHandler {
	return &AuthHandler{authClient: authClient}
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	var reqDTO LoginDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("authHandler.Login invalid json request", slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	grpcReq := &auth.LoginRequest{
		Email:    reqDTO.Email,
		Password: reqDTO.Password,
	}

	resp, err := ah.authClient.Login(r.Context(), grpcReq)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
		case codes.InvalidArgument, codes.Unauthenticated:
			response.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "invalid email or password")
		default:
			l.Error("authHandler.Login internal error", slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

type RegisterDTO struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (ah *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	var reqDTO RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("authHandler.Register invalid json request", slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	grpcReq := &auth.RegisterRequest{
		Email:    reqDTO.Email,
		Login:    reqDTO.Login,
		Name:     reqDTO.Name,
		Password: reqDTO.Password,
	}

	resp, err := ah.authClient.Register(r.Context(), grpcReq)
	if err != nil {
		switch status.Code(err) {
		case codes.AlreadyExists:
			response.Error(w, http.StatusConflict, "USER_ALREADY_EXISTS", "user with this email or login already exists")
		case codes.InvalidArgument:
			response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid input data")
		default:
			l.Error("authHandler.Register internal error", slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		}
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
