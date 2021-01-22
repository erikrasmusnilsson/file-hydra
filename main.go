package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-redis/redis"

	"./controllers"
	"./repositories"
	"github.com/julienschmidt/httprouter"
)

func main() {
	rc := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	_, err := rc.Ping(context.Background()).Result()

	if err != nil {
		panic(err)
	} else {
		log.Println("Connected to Redis.")
	}

	mux := httprouter.New()
	repo := repositories.NewSessionRepository(rc)
	sc := controllers.NewSessionController("public", repo)

	mux.POST("/sessions", sc.CreateSession)
	mux.GET("/sessions/:id", sc.GetSession)

	log.Println("Starting server.")
	http.ListenAndServe(":80", mux)
}
