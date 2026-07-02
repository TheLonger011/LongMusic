package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type PlaylistHandler struct {
	service *service.PlaylistService
}

type reqPlaylist struct {
	Name string `json:"name"`
}

type reqTrak struct {
	TrackID int64 `json:"track_id"`
}

func NewPlaylistHandler(service *service.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{service: service}
}

func (h *PlaylistHandler) Create(w http.ResponseWriter, r *http.Request) {
	var reqPlaylist reqPlaylist
	json.NewDecoder(r.Body).Decode(&reqPlaylist)

	userID := r.Context().Value("user_id").(int64)
	id, err := h.service.Create(userID, reqPlaylist.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})

}

func (h *PlaylistHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(int64)
	playlist, err := h.service.GetByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlist)

}

func (h *PlaylistHandler) AddTrack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	var reqTrak reqTrak
	json.NewDecoder(r.Body).Decode(&reqTrak)
	err := h.service.AddTrack(id, reqTrak.TrackID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (h *PlaylistHandler) GetTracks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	playlist, err := h.service.GetTracks(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return

	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(playlist)
}

func (h *PlaylistHandler) RemoveTrack(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is invalid"})
		return
	}

	var reqTrak reqTrak
	if err := json.NewDecoder(r.Body).Decode(&reqTrak); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is invalid"})
		return
	}

	err = h.service.RemoveTrack(id, reqTrak.TrackID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "id is invalid"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "track removed successfully"})
}
