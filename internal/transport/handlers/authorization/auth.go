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

// Login godoc
// @Summary Авторизация пользователя
// @Description Принимает email и пароль, возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginDTO true "Данные для входа"
// @Success 200 {object} auth.LoginResponse "вернёт access_token"
// @Failure 400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure 401 {object} response.ErrorResponse "Неверный логин или пароль"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /login [post]
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

// Register godoc
// @Summary Регистрация пользователя
// @Description Создает нового пользователя и возвращает JWT токен
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RegisterDTO true "Данные для регистрации"
// @Success 200 {object} auth.RegisterResponse "вернет access_token"
// @Failure 400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure 409 {object} response.ErrorResponse "Пользователь с таким email или логином уже существует"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /register [post]
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

type GetUsersInfoDTO struct {
	UserIDs []string `json:"user_ids"`
}

// GetUsersInfo godoc
// @Summary Получение информации о группе пользователей
// @Description Возвращает мапу с данными пользователей по переданному массиву ID
// @Tags users
// @Accept json
// @Produce json
// @Param request body GetUsersInfoDTO true "Массив UUID пользователей"
// @Success 200 {object} auth.GetUsersInfoResponse "Map пользователей (ключ - ID)"
// @Failure 400 {object} response.ErrorResponse "Неверный формат запроса"
// @Failure 404 {object} response.ErrorResponse "Пользователи не найдены"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/info [post]
func (ah *AuthHandler) GetUsersInfo(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	var reqDTO GetUsersInfoDTO
	if err := json.NewDecoder(r.Body).Decode(&reqDTO); err != nil {
		l.Error("authHandler.GetUsersInfo invalid json request", slog.String("error", err.Error()))
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	grpcReq := &auth.GetUsersInfoRequest{
		UserIds: reqDTO.UserIDs,
	}
	resp, err := ah.authClient.GetUsersInfo(r.Context(), grpcReq)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			l.Error("authHandler.GetUsersInfo user not found",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		default:
			l.Error("authHandler.GetUsersInfo internal error",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}
	response.JSON(w, http.StatusOK, resp)
}

type UserInfoByLoginDTO struct {
	Login string
}

// GetUserInfoByLogin godoc
// @Summary Получение публичного профиля пользователя
// @Description Ищет пользователя по уникальному логину (никнейму) и возвращает его публичные данные
// @Tags users
// @Produce json
// @Param login path string true "Логин пользователя (без @)"
// @Success 200 {object} auth.UserInfo "Публичные данные пользователя"
// @Failure 400 {object} response.ErrorResponse "Логин не передан"
// @Failure 404 {object} response.ErrorResponse "Пользователь не найден"
// @Failure 500 {object} response.ErrorResponse "Внутренняя ошибка сервера"
// @Router /users/{login} [get]
func (ah *AuthHandler) GetUserInfoByLogin(w http.ResponseWriter, r *http.Request) {
	l := logger.FromContext(r.Context())

	var reqDTO UserInfoByLoginDTO
	login := r.PathValue("login")
	if login == "" {
		l.Error("authHandler.GetUserInfoByLogin invalid login")
		response.Error(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request query")
		return
	}
	reqDTO.Login = login
	grpcReq := &auth.GetUserInfoByLoginRequest{
		Login: reqDTO.Login,
	}
	resp, err := ah.authClient.GetUserInfoByLogin(r.Context(), grpcReq)
	if err != nil {
		switch status.Code(err) {
		case codes.NotFound:
			l.Info("authHandler.GetUserInfoByLogin user not found",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", "user not found")
			return
		default:
			l.Error("authHandler.GetUserInfoByLogin internal error",
				slog.String("error", err.Error()))
			response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
			return
		}
	}
	response.JSON(w, http.StatusOK, resp)
}
