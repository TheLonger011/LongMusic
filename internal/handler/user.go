package handler

import (
	"encoding/json"
	"fmt"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
)

type publicProfileResponse struct {
	ID     int64  `json:"id"`
	Login  string `json:"login"`
	Avatar string `json:"avatar_url"`
}

type UserHandler struct {
	service *service.UserService
}

type reqUserName struct {
	Login string `json:"login"`
}

type profileResponse struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	Login        string `json:"login"`
	Role         string `json:"role"`
	Subscription string `json:"subscription"`
	AvatarURL    string `json:"avatar_url"`
}

func NewUserHandler(service *service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Login    string `json:"login"`
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("JSON decode error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Register request: email=%s, login=%s, password=%s\n", req.Email, req.Login, req.Password)

	id, err := h.service.Register(req.Email, req.Login, req.Password)
	if err != nil {
		fmt.Printf("Service register error: %v\n", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("User registered with ID: %d\n", id)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]int64{"id": id}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, user, err := h.service.Login(req.Login, req.Password)
	if err != nil {
		fmt.Printf("Login error: %v\n", err)
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
			AvatarURL:    user.Avatar,
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
		AvatarURL:    profile.Avatar,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *UserHandler) UpdateUsername(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	var reqUserName reqUserName

	if err := json.NewDecoder(r.Body).Decode(&reqUserName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateUsername(userID, reqUserName.Login); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"login": reqUserName.Login})
}

func (h *UserHandler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	file, _, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	fsPath := fmt.Sprintf("./uploads/avatar/%d.jpg", userID)
	urlPath := fmt.Sprintf("/uploads/avatar/%d.jpg", userID)

	if err := os.MkdirAll("./uploads/avatar", 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	dst, err := os.Create(fsPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	err = h.service.UpdateAvatar(userID, urlPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"avatar_url": urlPath})
}

func (h *UserHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	login := chi.URLParam(r, "login")
	user, err := h.service.GetPublicProfile(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(publicProfileResponse{
		ID:     user.ID,
		Login:  user.Login,
		Avatar: user.Avatar,
	})

}
