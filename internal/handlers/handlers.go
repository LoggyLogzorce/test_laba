package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"log"
	"masha_laba_3/internal/models"
	"masha_laba_3/internal/services"
	"net/http"
	"strconv"
)

func jsonEscape(i []byte) string {
	return base64.StdEncoding.EncodeToString(i)
}

type Handlers struct {
	services services.Services
}

func NewHandlers(services services.Services) *Handlers {
	return &Handlers{services: services}
}

// LoginResponse структура ответа авторизации
type LoginResponse struct {
	Success     bool   `json:"success"`
	Requires2FA bool   `json:"requires_2fa"`
	Message     string `json:"message,omitempty"`
}

func (h *Handlers) setSessionCookies(w http.ResponseWriter, userID uint, role models.UserRole) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    strconv.Itoa(int(userID)),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600, // 1 час
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "role",
		Value:    string(role),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   3600,
	})
}

func (h *Handlers) getCurrentUserID(r *http.Request) (uint, error) {
	sessionCookie, err := r.Cookie("session")
	if err != nil {
		return 0, err
	}

	userID, err := strconv.Atoi(sessionCookie.Value)
	if err != nil {
		return 0, err
	}

	return uint(userID), nil
}

func (h *Handlers) getCurrentUserLogin(r *http.Request) (string, error) {
	loginCookie, err := r.Cookie("auth_pending")
	if err != nil {
		return "", err
	}

	return loginCookie.Value, nil
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	login := r.FormValue("login")
	password := r.FormValue("password")

	result, err := h.services.Login(login, password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Success:     false,
			Requires2FA: false,
			Message:     err.Error(),
		})
		return
	}

	// Если требуется 2FA
	if result.TwoFA != "" {
		// Временная сессия для 2FA
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_pending",
			Value:    login,
			HttpOnly: true,
			MaxAge:   300, // 5 минут
		})
		json.NewEncoder(w).Encode(LoginResponse{
			Success:     true,
			Requires2FA: true,
			Message:     "2FA required",
		})
		return
	}

	// Установка основной сессии
	h.setSessionCookies(w, result.UserID, result.Role)
	json.NewEncoder(w).Encode(LoginResponse{
		Success:     true,
		Requires2FA: false,
		Message:     "OK",
	})
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var data *models.User
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.services.Register(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "session", Value: "", MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: "role", Value: "", MaxAge: -1})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *Handlers) TwoFAGenerate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.services.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	key, secret, err := h.services.Generate2FASecret(user.Login, user.Email)
	if err != nil {
		http.Error(w, "Failed to generate 2FA secret", http.StatusInternalServerError)
		return
	}

	img, err := key.Image(200, 200)
	if err != nil {
		http.Error(w, "Failed to generate QR code", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		http.Error(w, "Failed to encode QR code", http.StatusInternalServerError)
		return
	}

	qrCode := buf.Bytes()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"secret":  secret,
		"qr_code": "data:image/png;base64," + jsonEscape(qrCode),
	})
}

func (h *Handlers) SQLiDemo(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	tests, err := h.services.SQLi(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tests)
}

type TwoFAVerify struct {
	Code   string `json:"code"`
	Secret string `json:"secret"`
}

func (h *Handlers) TwoFAEnable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data *TwoFAVerify

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.services.Verify2FA(data.Secret, data.Code)
	if err != nil {
		http.Error(w, "неверный 2FA код", http.StatusBadRequest)
		return
	}

	err = h.services.Enable2FA(userID, data.Secret)
	if err != nil {
		http.Error(w, "Непредвиденная ошибка, 2fa не подключена", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) TwoFAStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.services.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	enabled := false
	if user.TwoFASecret != nil {
		enabled = true
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"enabled": enabled,
	})
}

func (h *Handlers) TwoFAVerify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var data *TwoFAVerify
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userLogin, err := h.getCurrentUserLogin(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.services.GetUserByLogin(userLogin)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if user.TwoFASecret == nil {
		http.Error(w, "SecretKey не установлен, подключите 2fa", http.StatusBadRequest)
		return
	}

	err = h.services.Verify2FA(*user.TwoFASecret, data.Code)
	if err != nil {
		http.Error(w, "неверный 2FA код", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "auth_pending", Value: "", Path: "/", MaxAge: -1})
	h.setSessionCookies(w, user.ID, user.Role)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

func (h *Handlers) GetTests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tests, err := h.services.GetTests()
	if err != nil {
		http.Error(w, "Failed to get tests", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tests)
}

func (h *Handlers) CreateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var test *models.Test
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		log.Println(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	test.UserID = userID

	err = h.services.CreateTest(test)
	if err != nil {
		http.Error(w, "Failed to create test", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *Handlers) UpdateTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var test *models.Test
	if err := json.NewDecoder(r.Body).Decode(&test); err != nil {
		log.Println(err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID, err := h.getCurrentUserID(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	test.UserID = userID

	err = h.services.UpdateTest(test)
	if err != nil {
		http.Error(w, "Failed to update test", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) DeleteTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err = h.services.DeleteTest(idInt)
	if err != nil {
		http.Error(w, "Failed to delete test", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) GetUserRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	role, err := r.Cookie("role")
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"role": role.Value,
	})
}

func (h *Handlers) GetCategory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	category, err := h.services.GetCategories()
	if err != nil {
		http.Error(w, "Failed to get categories", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

type QuestionsDTO struct {
	Test      *models.Test      `json:"test"`
	Questions []models.Question `json:"questions"`
}

func (h *Handlers) GetQuestionsByTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	questions, err := h.services.GetQuestionsByTest(uint(idUint))
	if err != nil {
		http.Error(w, "Failed to get questions", http.StatusInternalServerError)
		return
	}

	test, err := h.services.GetTestByID(uint(idUint))
	if err != nil {
		http.Error(w, "Failed to get test", http.StatusInternalServerError)
		return
	}

	result := QuestionsDTO{
		Test:      test,
		Questions: questions,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *Handlers) GetOptionsByQuestion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	answerOptions, err := h.services.GetAnswerOptionsByQuestionID(uint(idUint))
	if err != nil {
		http.Error(w, "Failed to get answer options", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(answerOptions)
}
