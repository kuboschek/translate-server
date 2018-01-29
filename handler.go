package main

import (
	"bytes"
	"golang.org/x/text/language"
	"io"
	"log"
	"net/http"
	"time"
)

var services = &[]TranslateService{
	TestProvider{
		failing: false,
		delay:   time.Second * 2,
	},
}

// writeResponse sets appropriate headers, then writes the translated string to the
// http.ResponseWriter
func writeResponse(w http.ResponseWriter, targetLanguage, targetPhrase string) {
	headers := w.Header()

	headers.Set("Content-Type", "text/plain")
	headers.Set("Content-Language", targetLanguage)
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, targetPhrase)
}

func translateHandler(response http.ResponseWriter, request *http.Request) {
	// Disallow anything but POST requests
	if request.Method != http.MethodPost {
		response.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(response, "Only POST requests are allowed")
	}

	// Get content language (Content-Language header)
	contentLang := request.Header.Get("Content-Language")
	tags, _, err := language.ParseAcceptLanguage(request.Header.Get("Accept-Language"))

	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(err.Error()))
		return
	}

	if contentLang == "" {
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
	targetLanguage := tags[0].String()

	// Check for a cached response
	cached := GetTranslation(targetLanguage, givenPhrase)
	if cached != nil {
		writeResponse(response, targetLanguage, *cached)
		return
	}

	serviceResponse := make(chan TranslateResult)

	// Go through all the services in order - return the first successful result
	for _, svc := range *services {
		go svc.GetTranslation(givenPhrase, contentLang, targetLanguage, &serviceResponse)

		// Wait for the response from the service
		result := <-serviceResponse

		if result.err == nil {
			AddTranslation(result.targetLang, result.givenPhrase, result.translatedPhrase)
			writeResponse(response, result.targetLang, result.translatedPhrase)
			return
		}

		log.Printf("failed to fetch translations: %v", err)
	}

	log.Printf("all services failed to translate \"%v\" (%v -> %v)", givenPhrase, contentLang, targetLanguage)

	// At this point, we've run out of services to try - so we fail hard, and respond with an error
	response.WriteHeader(http.StatusInternalServerError)
}
