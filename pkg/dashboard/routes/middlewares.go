package routes

import (
	"log"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := strategy.Authenticate(r.Context(), r)
		if err != nil {
			log.Println("[ERROR] ", err)
			code := http.StatusUnauthorized
			http.Error(w, http.StatusText(code), code)
			return
		}
		log.Printf("[INFO] User %s Authenticated\n", user.GetUserName())
		next.ServeHTTP(w, r)
	})
}
