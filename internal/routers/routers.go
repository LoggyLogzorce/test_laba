package routers

import (
	"masha_laba_3/internal/handlers"
	"masha_laba_3/internal/middleware"
	"net/http"
)

// SetupRoutes настраивает все маршруты для приложения
func SetupRoutes(h *handlers.Handlers) *http.ServeMux {
	mux := http.NewServeMux()

	// Pages
	mux.HandleFunc("/login", h.LoginPage)
	mux.HandleFunc("/register", h.RegisterPage)
	mux.HandleFunc("/2fa", h.TwoFAPage)
	mux.HandleFunc("/2fa-setup", h.TwoFASetupPage)
	mux.HandleFunc("/tests", middleware.AuthRequired(h.TestsPage))
	mux.HandleFunc("/test", middleware.AuthRequired(h.TestPage))

	// Demo
	mux.HandleFunc("/demo/sqli", h.SQLiDemo)

	// API Tests
	mux.HandleFunc("/api/tests", middleware.AuthRequired(h.GetTests))
	mux.HandleFunc("/api/test/create", middleware.AuthRequired(middleware.RBAC("admin", "tester")(h.CreateTest)))
	mux.HandleFunc("/api/test/update", middleware.AuthRequired(middleware.RBAC("admin", "tester")(h.UpdateTest)))
	mux.HandleFunc("/api/test/delete/{id}", middleware.AuthRequired(middleware.RBAC("admin")(h.DeleteTest)))
	mux.HandleFunc("/api/categories", middleware.AuthRequired(h.GetCategory))
	mux.HandleFunc("/api/questions/test/{id}", middleware.AuthRequired(h.GetQuestionsByTest))
	mux.HandleFunc("/api/question/{id}/options", middleware.AuthRequired(h.GetOptionsByQuestion))

	// API Auth
	mux.HandleFunc("/api/login", h.Login)
	mux.HandleFunc("/api/register", h.Register)
	mux.HandleFunc("/logout", h.Logout)

	// 2FA
	mux.HandleFunc("/api/2fa/generate", middleware.AuthRequired(h.TwoFAGenerate))
	mux.HandleFunc("/api/2fa/status", middleware.AuthRequired(h.TwoFAStatus))
	mux.HandleFunc("/api/2fa/enable", h.TwoFAEnable)
	mux.HandleFunc("/api/2fa/verify", h.TwoFAVerify)

	// API Users
	mux.HandleFunc("/api/get_role", middleware.AuthRequired(h.GetUserRole))

	return mux
}
