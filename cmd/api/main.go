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

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	trackRepo := repository.NewTrackRepository(db)
	trackService := service.NewTrackService(trackRepo)
	trackHandler := handler.NewTrackHandler(trackService)

	r.Post("/register", userHandler.Register)
	r.Post("/login", userHandler.Login)

	r.Group(func(r chi.Router) {
		r.Use(handler.AuthMiddleware)
		r.Get("/me", userHandler.Me)
		r.Post("/tracks", trackHandler.Upload)
		r.Get("/tracks/{id}", trackHandler.GetByID)
	})

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "ok"}`))
	})

	http.ListenAndServe(":8080", r)
}
