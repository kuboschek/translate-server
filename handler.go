package main

import (
	"bytes"
	"github.com/kuboschek/translate-server/cache"
	"github.com/kuboschek/translate-server/upstream"
	"golang.org/x/text/language"
	"io"
	"log"
	"net/http"
	"time"
)

const timeout = time.Second * 5

// TranslateHandler is an HTTP handler that proxies translation requests to upstream providers.
//
type TranslateHandler struct {
	Services []upstream.Service
	Cache    cache.Cache
}

// writeSuccess sets appropriate headers, then writes the translated string to the ResponseWriter
func writeSuccess(w http.ResponseWriter, targetLanguage language.Tag, targetPhrase string) {
	headers := w.Header()

	headers.Set("Content-Type", "text/plain")
	headers.Set("Content-Language", targetLanguage.String())
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, targetPhrase)
}

// moveToBack moves a upstream service reference to the back of the list
func (h TranslateHandler) moveToBack(index int) {
	var movedService = h.Services[index]

	copy(h.Services[index:], h.Services[index+1:])
	h.Services[len(h.Services)-1] = nil // set to the zero value
	h.Services = append(h.Services[:len(h.Services)-1], movedService)
}

func (h TranslateHandler) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	// Disallow anything but POST requests
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(response, "Only POST requests are allowed")
		return
	}

	// Get and parse target language (Accept-Language header)
	tags, _, err := language.ParseAcceptLanguage(request.Header.Get("Accept-Language"))

	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	// Get given language (Content-Language header)
	contentLanguage, err := language.Parse(request.Header.Get("Content-Language"))

	if err != nil {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("No Content-Language header specified\n"))
		return
	}

	// Get target language (Accept-Language header)
	if len(tags) < 1 {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("No Accept-Language header specified\n"))
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(request.Body)
	givenPhrase := buf.String()

	targetLanguage := tags[0]

	// Check for a cached response
	if h.Cache != nil {
		cached, err := h.Cache.Get(givenPhrase, targetLanguage)
		if err == nil {
			writeSuccess(response, targetLanguage, cached)
			return
		}
	}

	serviceResponse := make(chan upstream.Result)

	// Go through all the services in order - return the first successful result
	for index, svc := range h.Services {
		go svc.Translate(givenPhrase, contentLanguage, targetLanguage, &serviceResponse)

		// Wait for the response from the service for a specified time
		select {
		case result := <-serviceResponse:
			if result.Error == nil {
				if h.Cache != nil {
					h.Cache.Put(result.GivenPhrase, result.TargetLang, result.TranslatedPhrase)
				}

				writeSuccess(response, result.TargetLang, result.TranslatedPhrase)
				return
			}
			// Move the failing service to the end of the list
			h.moveToBack(index)
			log.Printf("failed to fetch translations: %v", result.Error)

			break

		case <-time.After(timeout):
			log.Printf("upstream service timed out after: %v", timeout)
			break
		}
	}

	log.Printf("all services failed to translate \"%v\" (%v -> %v)", givenPhrase, contentLanguage, targetLanguage)

	// At this point, we've run out of services to try - so we fail hard, and respond with an error
	response.WriteHeader(http.StatusBadGateway)
	io.WriteString(response, "All upstream services failed to translate.")
}
