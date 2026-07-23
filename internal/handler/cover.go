package handler

import (
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/bogem/id3v2/v2"
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

type CoverHandler struct {
	service *service.TrackService
}

func NewCoverHandler(service *service.TrackService) *CoverHandler {
	return &CoverHandler{service: service}
}

func (h *CoverHandler) GetCover(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	track, err := h.service.GetByID(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tag, err := id3v2.Open(track.FilePath, id3v2.Options{Parse: true})
	if err != nil {
		http.Error(w, "cover not found", http.StatusInternalServerError)
		return
	}
	defer tag.Close()
	frames := tag.GetFrames(tag.CommonID("Attached picture"))
	if len(frames) > 0 {
		pic := frames[0].(id3v2.PictureFrame)
		w.Header().Set("Content-Type", pic.MimeType)
		w.Write(pic.Picture)
	}

}
