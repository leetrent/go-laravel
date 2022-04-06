package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"myapp/data"
	"net/http"
	"time"
)

func (h *Handlers) UserLogin(w http.ResponseWriter, r *http.Request) {
	err := h.App.Render.Page(w, r, "login", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println(err)
	}
}

func (h *Handlers) PostUserLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		w.Write([]byte("Error validating password"))
		return
	}

	if !matches {
		w.Write([]byte("Invalid password"))
		return
	}

	// DID USER CHECK 'REMEMBER ME'?
	if r.Form.Get("remember") == "remember" {
		randomString := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomString))
		if err != nil {
			h.App.ErrorLog.Println("[auth-handlers.go][PostUserLogin][hasher.Write] => (error):", err)
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}
		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RememberToken{}
		err = rm.InsertToken(user.ID, sha)
		if err != nil {
			h.App.ErrorLog.Println("[auth-handlers.go][PostUserLogin][InsertToken] => (error):", err)
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}
		// SET COOKIE
		expire := time.Now().Add(365 * 24 * 60 * 60 * time.Second)
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315350000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		// SAVE HASH IN SESSION
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}

	h.App.Session.Put(r.Context(), "userID", user.ID)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) UserLogout(w http.ResponseWriter, r *http.Request) {
	// DELETE REMEMBER TOKEN IF IT EXISTS
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := data.RememberToken{}
		_ = rt.DeleteToken(h.App.Session.GetString(r.Context(), "remember_token"))
	}

	// DELETE COOKIE
	deleteCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &deleteCookie)

	h.App.Session.RenewToken(r.Context())
	h.App.Session.Remove(r.Context(), "userID")
	h.App.Session.Remove(r.Context(), "remember_token")
	h.App.Session.Destroy(r.Context())
	h.App.Session.RenewToken(r.Context())

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
