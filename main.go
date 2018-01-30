package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

const (
	cachePath = "cache.json"
	serveAddr = "127.0.0.1"
)

var theHandler = make(TranslateHandler, 0)

// Tries to load the cache from a file
func init() {
	file, err := os.OpenFile(cachePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		log.Printf("could not load cache file: %v", err)
		return
	}

	if err = LoadCache(file); err != nil {
		log.Printf("could not decode cache file: %v", err)
	}

	gkey := os.Getenv("GOOGLE_API_KEY")
	if gkey != "" {
		theHandler = append(theHandler, GcloudProvider{
			apiKey: gkey,
		})
	}

	if len(theHandler) == 0 {
		err := errors.New("no translation backends active, exiting")
		log.Fatal(err)
	}
}

func main() {
	http.Handle("/", theHandler)

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
	if file, err := os.OpenFile(cachePath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
		log.Printf("failed to open cache file: %v", err)
	} else {
		err = StoreCache(file)
		if err != nil {
			log.Printf("failed to save cache to file: %v", err)
		}
	}

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
	defer cancel()
	err := srv.Shutdown(ctx)

	if err != nil {
		log.Printf("error shutting down: %v", err)
	}
}
