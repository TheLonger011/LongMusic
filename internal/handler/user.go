package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"net/http"
)

type UserHandler struct {
	service *service.UserService
}

type profileResponse struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Login        string `json:"login"`
	Role         string `json:"role"`
	Subscription string `json:"subscription"`
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
	token, user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"token": token,
		"user": profileResponse{
			ID:           user.ID,
			Email:        user.Email,
			Login:        user.Login,
			Role:         user.Role,
			Subscription: user.Subscription,
		},
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"user_id": userID})
}

func (h *UserHandler) Profile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	profile, err := h.service.GetProfile(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	resp := profileResponse{
		ID:           profile.ID,
		Email:        profile.Email,
		Login:        profile.Login,
		Role:         profile.Role,
		Subscription: profile.Subscription,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
