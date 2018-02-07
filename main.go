package main

import (
	"context"
	"errors"
	"github.com/kuboschek/translate-server/cache"
	"github.com/kuboschek/translate-server/upstream"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var translateHandler TranslateHandler

// init adcs translation handlers based on the environment variables present
func init() {
	translateHandler = TranslateHandler{
		Cache: cache.Memory,
	}

	googleKey := os.Getenv("GOOGLE_API_KEY")
	if googleKey != "" {
		translateHandler.Services = append(translateHandler.Services, upstream.GoogleProvider{
			Key: googleKey,
		})
	}

	if len(translateHandler.Services) == 0 {
		err := errors.New("no translation backends active, exiting")
		log.Fatal(err)
	}
}

func main() {
	http.Handle("/", translateHandler)

	// Setting timeouts here to mitigate certain Denial-of-Service attacks
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      http.DefaultServeMux,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, os.Kill)

	log.Printf("Listening on %v", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	<-gracefulStop
	log.Println("Shutting down")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
	defer cancel()
	err := srv.Shutdown(ctx)

	if err != nil {
		log.Printf("error shutting down: %v", err)
	}
}
