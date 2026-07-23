package handler

import (
	"encoding/json"
	"github.com/TheLonger011/LongMusic/internal/domain"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type ChartHandler struct {
	service *service.ChartService
}

func NewChartHandler(service *service.ChartService) *ChartHandler {
	return &ChartHandler{service: service}
}

func (s *ChartHandler) GetChart(w http.ResponseWriter, r *http.Request) {
	chartId := chi.URLParam(r, "id")

	var tracks []domain.Track
	var err error

	switch chartId {
	case "top50":
		tracks, err = s.service.GetTopTracks(50)
	case "top100":
		tracks, err = s.service.GetTopTracks(100)
	case "today":
		tracks, err = s.service.GetTopTracksToday(50)
	default:
		http.Error(w, "chart not found", http.StatusNotFound)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tracks)
}
