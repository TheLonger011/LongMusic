package main

import (
	"fmt"
	"github.com/TheLonger011/LongMusic/internal/handler"
	"github.com/TheLonger011/LongMusic/internal/repository"
	"github.com/TheLonger011/LongMusic/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

func main() {
	db, err := repository.NewPostgresDB("postgres://postgres:181818@localhost:5432/longmusic?sslmode=disable")
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
	r.Get("/api/playlists/public/search", playlistHandler.SearchPublic)
	r.Get("/api/charts/{id}", chartHandler.GetChart)
	r.Get("/api/tracks", trackHandler.GetAll)
	r.Get("/api/tracks/search", trackHandler.Search)
	r.Get("/api/tracks/{id}", trackHandler.GetByID)
	r.Get("/api/tracks/{id}/stream", trackHandler.Stream)
	r.Get("/api/stream/{id}", trackHandler.Stream)
	r.Get("/api/users/{login}/profile", userHandler.GetPublicProfile)

	r.Group(func(r chi.Router) {
		r.Use(handler.OptionalAuthMiddleware)
		r.Get("/api/playlists/{id}", playlistHandler.GetByID)
		r.Get("/api/playlists/{id}/tracks", playlistHandler.GetTracks)
	})

	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware)
		r.Get("/api/me", userHandler.Me)
		r.Post("/api/tracks", trackHandler.Upload)
		r.Post("/api/plays", playHandler.RecordPlay)
		r.Post("/api/history", playHandler.RecordPlay)
		r.Get("/api/plays", playHandler.GetHistory)
		r.Post("/api/playlists", playlistHandler.Create)
		r.Get("/api/playlists", playlistHandler.GetByUserID)
		r.Post("/api/playlists/{id}/tracks", playlistHandler.AddTrack)
		r.Get("/api/profile", userHandler.Profile)
		r.Get("/api/history", playHandler.History)
		r.Delete("/api/playlists/{id}/tracks", playlistHandler.RemoveTrack)
		r.Post("/api/favorites", favoriteHandler.Add)
		r.Delete("/api/favorites/{id}", favoriteHandler.Remove)
		r.Get("/api/favorites", favoriteHandler.GetByUserID)
		r.Patch("/api/profile/username", userHandler.UpdateUsername)
		r.Post("/api/profile/avatar", userHandler.UpdateAvatar)
		r.Patch("/api/playlists/{id}/publish", playlistHandler.SetPublic)

	})

	r.Post("/api/now-playing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Post("/api/now-playing/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.Delete("/api/now-playing", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	fs := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	uploadsFs := http.FileServer(http.Dir("uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", uploadsFs))

	r.Get("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/sw.js")
	})
	r.Get("/player.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/player.js")
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.Get("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})
	r.Get("/login.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/login.html")
	})
	r.Get("/register.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/register.html")
	})
	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/index.html")
	})

	seedTracks(db)
	fmt.Println("start")

	http.ListenAndServe(":8080", r)

}
