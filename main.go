package main

import (
	"fmt"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"os"
)

const (
	cachePath = "cache.gob"
	serveAddr = "127.0.0.1"
)

var services []TranslateService

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
}

func translateHandler(response http.ResponseWriter, request *http.Request) {
	// Get content language (Content-Language header)
	contentLang := request.Header.Get("Content-Language")
	tags, _, err := language.ParseAcceptLanguage(request.Header.Get("Accept-Language"))

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
	}

	if contentLang == "" {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("No content language specified"))
	}

	// Get target language (Accept-Language header)
	if len(tags) < 1 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("No target language specified"))
	} else {
		// TODO(kuboschek): Call the first translate service, wait for results.
		// If it fails, call the next in the row
	}
}

func main() {
	fmt.Println("Hello World!")

	http.HandleFunc("/translate", translateHandler)
	http.ListenAndServe(serveAddr, http.DefaultServeMux)
}
