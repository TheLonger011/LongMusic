package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type FavoriteHandler struct {
	service *service.FavoriteService
}

func NewFavoriteHandler(service *service.FavoriteService) *FavoriteHandler {
	return &FavoriteHandler{service: service}
}

func (s *FavoriteHandler) Add(w http.ResponseWriter, r *http.Request) {
	var req req
	json.NewDecoder(r.Body).Decode(&req)
	userID := r.Context().Value("user_id").(int64)

	err := s.service.Add(userID, req.TrackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": req.TrackID})
}

func (s *FavoriteHandler) Remove(w http.ResponseWriter, r *http.Request) {

	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value("user_id").(int64)
	err = s.service.Remove(userID, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (s *FavoriteHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	favorite, err := s.service.GetByUserID(userID)
	if err != nil {

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(favorite)

}
