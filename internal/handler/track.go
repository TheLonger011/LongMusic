package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type TrackHandler struct {
	service *service.TrackService
}

func NewTrackHandler(service *service.TrackService) *TrackHandler {
	return &TrackHandler{service: service}
}

func (h *TrackHandler) Upload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	artist := r.FormValue("artist")
	album := r.FormValue("album")
	duration, _ := strconv.ParseInt(r.FormValue("duration"), 10, 64)

	id, err := h.service.Upload(name, artist, album, duration, file, header.Filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")

	err = json.NewEncoder(w).Encode(map[string]int64{"id": id})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}

func (h *TrackHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	track, err := h.service.GetByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("content-type", "application/json")
	err = json.NewEncoder(w).Encode(track)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
