package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	testContent = "testContent"
	testLang    = "testLang"
)

func TestWriteResponse(t *testing.T) {
	rr := httptest.NewRecorder()
	writeResponse(rr, testLang, testContent)

	if rr.Body.String() != testContent {
		t.Error("writeResponse should write exactly the given phrase to the output.")
	}

	if rr.Header().Get("Content-Language") != testLang {
		t.Error("writeResponse should set the Content-Language header.")
	}

	if rr.Header().Get("Content-Type") != "text/plain" {
		t.Error("writeResponse should set the Content-Type header.")
	}
}
func TestTranslateHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	translateHandler(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("translateHandler should not allow %v requests.", http.MethodGet)
	}
}

func TestTranslateHandler2(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en;q=1.2.3")
	rr := httptest.NewRecorder()

	translateHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Error("translateHandler should reject malformed Accept-Language headers.")
	}
}

func TestTranslateHandler3(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Accept-Language", "en,fr")
	rr := httptest.NewRecorder()

	translateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Content-Language headers.")
	}
}

func TestTranslateHandler4(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Language", "fr")
	rr := httptest.NewRecorder()

	translateHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Error("translateHandler should reject request with missing Accept-Language headers.")
	}
}

func TestTranslateHandler5(t *testing.T) {
	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	translateHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Error("translateHandler should accept requests with appropriate headers.")
	}
}

func TestTranslateHandler6(t *testing.T) {
	var serviceBackup = services

	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	// Use this to load the package, so we can then reset the services list
	translateHandler(rr, req)

	services = &[]TranslateService{}
	rr = httptest.NewRecorder()
	translateHandler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Error("translateHandler should reject requests when all services fail.")
	}

	services = serviceBackup
}

func TestTranslateHandler7(t *testing.T) {
	var serviceBackup = services
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	content := bytes.NewBufferString("Guten Morgen.")
	req := httptest.NewRequest(http.MethodPost, "/", content)
	req.Header.Set("Accept-Language", "en,fr")
	req.Header.Set("Content-Language", "de")
	rr := httptest.NewRecorder()

	// Use this to load the package, so we can then reset the services list
	translateHandler(rr, req)

	services = &[]TranslateService{
		TestProvider{
			delay:   0,
			failing: true,
		},
	}
	rr = httptest.NewRecorder()
	translateHandler(rr, req)

	if buf.Len() == 0 {
		t.Error("translateHandler should log when a service fails.")
	}

	services = serviceBackup
}
