package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
	"os"
	"strconv"
	"time"
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

func (h *TrackHandler) Stream(w http.ResponseWriter, r *http.Request) {
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

	file, err := os.Open(track.FilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "audio/mpeg")
	http.ServeContent(w, r, track.FilePath, time.Time{}, file)
}

func (h *TrackHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.service.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}

func (h *TrackHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	tracks, err := h.service.Search(q)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}

func (h *TrackHandler) GetArtists(w http.ResponseWriter, r *http.Request) {
	tracks, err := h.service.GetArtists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)

}

func (h *TrackHandler) GetArtistTracks(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tracks, err := h.service.GetArtistsTracks(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	artistName := ""
	if len(tracks) > 0 {
		artistName = tracks[0].Artist
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"artist": map[string]string{"name": artistName},
		"tracks": tracks,
	})

}
