package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"net/http"
)

type PlayHandler struct {
	service *service.PlayService
}

type req struct {
	TrackID int64 `json:"track_id"`
}

func NewPlayHandler(service *service.PlayService) *PlayHandler {
	return &PlayHandler{service: service}
}

func (h *PlayHandler) RecordPlay(w http.ResponseWriter, r *http.Request) {
	var req req
	json.NewDecoder(r.Body).Decode(&req)

	userID := r.Context().Value("user_id").(int64)

	err := h.service.RecordPlay(userID, req.TrackID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
}

func (h *PlayHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)

	plays, err := h.service.GetHistory(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plays)
}

func (h *PlayHandler) History(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	tracks, err := h.service.GetHistoryWithTracks(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}
