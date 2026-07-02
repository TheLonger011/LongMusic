package main

import (
	"github.com/TheLonger011/LongMusic/internal/handler"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	db, err := repository.NewPostgresDB("postgres://postgres:181818@db:5432/longmusic?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	redisClient := repository.NewRedisClient()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	trackRepo := repository.NewTrackRepository(db)
	trackService := service.NewTrackService(trackRepo, redisClient)
	trackHandler := handler.NewTrackHandler(trackService)

	playRepo := repository.NewPlayRepository(db)
	playService := service.NewPlayService(playRepo)
	playHandler := handler.NewPlayHandler(playService)

	playlistRepo := repository.NewPlaylistRepository(db)
	playlistService := service.NewPlaylistService(playlistRepo)
	playlistHandler := handler.NewPlaylistHandler(playlistService)

	coverHandler := handler.NewCoverHandler(trackService)

	favoriteRepo := repository.NewFavoriteRepository(db)
	favoriteService := service.NewFavoriteService(favoriteRepo)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)

	chartRepo := repository.NewChartRepository(db)
	chartService := service.NewChartService(chartRepo)
	chartHandler := handler.NewChartHandler(chartService)

	r.Post("/api/auth/register", userHandler.Register)
	r.Post("/api/auth/login", userHandler.Login)
	r.Get("/api/cover/{id}", coverHandler.GetCover)
	r.Get("/api/artists", trackHandler.GetArtists)
	r.Get("/api/artists/{id}/tracks", trackHandler.GetArtistTracks)

	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware)
		r.Get("/api/me", userHandler.Me)
		r.Post("/api/tracks", trackHandler.Upload)
		r.Get("/api/tracks/search", trackHandler.Search)
		r.Get("/api/tracks/{id}", trackHandler.GetByID)
		r.Get("/api/tracks/{id}/stream", trackHandler.Stream)
		r.Get("/api/tracks", trackHandler.GetAll)
		r.Post("/api/plays", playHandler.RecordPlay)
		r.Get("/api/plays", playHandler.GetHistory)
		r.Post("/api/playlists", playlistHandler.Create)
		r.Get("/api/playlists", playlistHandler.GetByUserID)
		r.Post("/api/playlists/{id}/tracks", playlistHandler.AddTrack)
		r.Get("/api/playlists/{id}/tracks", playlistHandler.GetTracks)
		r.Get("/api/stream/{id}", trackHandler.Stream)
		r.Get("/api/profile", userHandler.Profile)
		r.Get("/api/history", playHandler.History)
		r.Delete("/api/playlists/{id}/tracks", playlistHandler.RemoveTrack)
		r.Post("/api/favorites", favoriteHandler.Add)
		r.Delete("/api/favorites/{id}", favoriteHandler.Remove)
		r.Get("/api/favorites", favoriteHandler.GetByUserID)
		r.Get("/api/charts/{id}", chartHandler.GetChart)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	http.ListenAndServe(":8080", r)
}
