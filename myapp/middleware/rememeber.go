package middleware

import (
	"fmt"
	"myapp/data"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (m *Middleware) CheckRememebr(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.App.Session.Exists(r.Context(), "userID") {
			// USER IS NOT LOGGED IN
			cookie, err := r.Cookie(fmt.Sprintf("_%s_remember", m.App.AppName))
			if err != nil {
				// NO COOKIE SO ON TO THE NEXT MIDDLEWARE
				next.ServeHTTP(w, r)
			} else {
				// CHECK FOUND COOKIE
				key := cookie.Value
				var u data.User
				if len(key) > 0 {
					// COOKIE HAS SOME DATA SO VALIDATE IT
					split := strings.Split(key, "|")
					uid, hash := split[0], split[1]
					id, err := strconv.Atoi(uid)
					if err != nil {
						fmt.Println("[remember.go][CheckRemember][strconv.Atoi] => (error):", err)
					}
					validHash := u.CheckForRememberToken(id, hash)
					if !validHash {
						m.deleteRememberCookie(w, r)
						m.App.Session.Put(r.Context(), "error", "You have been logged out from another device.")
						next.ServeHTTP(w, r)
					} else {
						// HASH IS VALID SO LOG THE USER IN
						user, err := u.GetByID(id)
						if err != nil {
							fmt.Println("[remember.go][CheckRemember][user.GetByID] => (error):", err)
						} else {
							m.App.Session.Put(r.Context(), "userID", user.ID)
							m.App.Session.Put(r.Context(), "remember_token", hash)
							next.ServeHTTP(w, r)
						}
					}
				} else {
					// KEY LENGTH IS ZERO
					// PROBABLY A LEFTOVER COOKIE (USER HAS NOT CLOSED BROWSER TAB)
					m.deleteRememberCookie(w, r)
					next.ServeHTTP(w, r)
				}
			}
		} else {
			// USER IS LOGGED IN
			next.ServeHTTP(w, r)
		}
	})
}

func (m *Middleware) deleteRememberCookie(w http.ResponseWriter, r *http.Request) {
	_ = m.App.Session.RenewToken(r.Context())

	// DELETE COOKIE
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", m.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   m.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   m.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, &newCookie)

	// LOG THE USER OUT
	m.App.Session.Remove(r.Context(), "userID")
	m.App.Session.Destroy(r.Context())

	_ = m.App.Session.RenewToken(r.Context())
}
