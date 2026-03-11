package middleware

import "net/http"

func AuthRequired(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		// Валидация сессии здесь
		next.ServeHTTP(w, r)
	}
}

func RBAC(allowedRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			c, err := r.Cookie("role")
			if err != nil {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			for _, role := range allowedRoles {
				if c.Value == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
		}
	}
}
