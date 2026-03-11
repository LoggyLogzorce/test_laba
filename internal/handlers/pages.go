package handlers

import "net/http"

func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/login.html")
}

func (h *Handlers) RegisterPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/register.html")
}

func (h *Handlers) TwoFAPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/2fa.html")
}

func (h *Handlers) TwoFASetupPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/2fa-setup.html")
}

func (h *Handlers) TestsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/tests.html")
}

func (h *Handlers) TestPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/templates/questions.html")
}
