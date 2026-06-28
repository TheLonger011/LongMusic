package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"net/http"
)

type UserHandler struct {
	service *service.UserService
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int64{"id": id}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": user}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"user_id": userID})
}
