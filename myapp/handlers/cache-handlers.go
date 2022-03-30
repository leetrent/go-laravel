package handlers

import (
	"fmt"
	"net/http"
)

func (h *Handlers) ShowCachePage(w http.ResponseWriter, r *http.Request) {
	err := h.render(w, r, "cache", nil, nil)
	if err != nil {
		h.App.ErrorLog.Println("[ShowCachePage] => (h.render): ", err)
		h.App.Error500(w, r)
		return
	}
}

func (h *Handlers) SaveInCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		CSRF  string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("[SaveInCache] => (ReadJSON): ", err)
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Set(userInput.Name, userInput.Value)
	if err != nil {
		h.App.ErrorLog.Println("[SaveInCache] => (Cache.Set): ", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}

	resp.Error = false
	resp.Message = "Saved in cache"

	err = h.App.WriteJSON(w, http.StatusCreated, resp)
	if err != nil {
		h.App.ErrorLog.Println("[SaveInCache] => (WriteJSON): ", err)
		h.App.Error500(w, r)
		return
	}
}

func (h *Handlers) GetFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("[GetFromCache] => (ReadJSON): ", err)
		h.App.Error500(w, r)
		return
	}

	var inCache = true
	var msg string

	fromCache, err := h.App.Cache.Get(userInput.Name)
	if err != nil {
		inCache = false
		msg = fmt.Sprintf("'%s' not found in cache!", userInput.Name)
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
		Value   string `json:"value"`
	}

	if inCache {
		resp.Error = false
		resp.Message = "Success"
		resp.Value = fromCache.(string)
	} else {
		resp.Error = true
		resp.Message = msg
	}

	err = h.App.WriteJSON(w, http.StatusCreated, resp)
	if err != nil {
		h.App.ErrorLog.Println("[SaveInCache] => (WriteJSON): ", err)
		h.App.Error500(w, r)
		return
	}
}

func (h *Handlers) DeleteFromCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		Name string `json:"name"`
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("[DeleteFromCache] => (ReadJSON): ", err)
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Forget(userInput.Name)
	if err != nil {
		h.App.ErrorLog.Println("[DeleteFromCache] => (Cache.Forget): ", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Deleted from cache (if it existed)"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)
}

func (h *Handlers) EmptyCache(w http.ResponseWriter, r *http.Request) {
	var userInput struct {
		CSRF string `json:"csrf_token"`
	}

	err := h.App.ReadJSON(w, r, &userInput)
	if err != nil {
		h.App.ErrorLog.Println("[EmptyCache] => (ReadJSON): ", err)
		h.App.Error500(w, r)
		return
	}

	err = h.App.Cache.Empty()
	if err != nil {
		h.App.ErrorLog.Println("[EmptyCache] => (Cache.Empty): ", err)
		h.App.Error500(w, r)
		return
	}

	var resp struct {
		Error   bool   `json:"error"`
		Message string `json:"message"`
	}
	resp.Error = false
	resp.Message = "Emptied cache!"

	_ = h.App.WriteJSON(w, http.StatusCreated, resp)

}
