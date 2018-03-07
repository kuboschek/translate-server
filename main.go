// translate-server is a caching text translation server.
// Supports pluggable upstream and cache backends.
// Fails over when upstreams return errors or take too long.
package main

import (
	"context"
	"errors"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/kuboschek/translate-server/cache"
	"github.com/kuboschek/translate-server/upstream"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

var (
	translateHandler TranslateHandler
	tokenKey         string
)

// init adds translation handlers based on the environment variables present
func init() {
	translateHandler = TranslateHandler{
		Cache: cache.Memory,
	}

	// Enable the Google backend if a key is given
	googleKey := os.Getenv("GOOGLE_API_KEY")
	if googleKey != "" {
		translateHandler.Services = append(translateHandler.Services, upstream.Google{
			Key: googleKey,
		})
	}

	// Enable the Bing backend if a key is given
	bingKey := os.Getenv("BING_API_KEY")
	if bingKey != "" {
		translateHandler.Services = append(translateHandler.Services, upstream.Azure{
			ServiceKey: bingKey,
		})
	}

	// This is useful for testing, enables a failing mock backend
	enableMock := os.Getenv("ENABLE_MOCK")
	if enableMock != "" {
		translateHandler.Services = append(translateHandler.Services, upstream.Mock{
			Failing: true,
		})
	}

	// This is the secret key used to sign JSON Web Tokens
	tokenKey = os.Getenv("SECRET_KEY")
	if tokenKey == "" {
		// TODO Replace by generated character string
		tokenKey = "uihesrioesjrjoiseros"
	}

	// If nothing is enabled, all requests would immediately fail
	// Shutting down immediately makes the configuration error more
	// observable
	if len(translateHandler.Services) == 0 {
		err := errors.New("no translation backends active, exiting")
		log.Fatal(err)
	}
}

func main() {
	http.Handle("/", translateHandler)

	// This adds simple authentication to the service.
	// Any bearer of a valid token may translate as much as they desire.
	tokenMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(tokenKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
	handler := tokenMiddleware.Handler(http.DefaultServeMux)

	// Setting timeouts here to mitigate certain Denial-of-Service attacks
	srv := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      handler,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}

	// Setting up a signal listener to allow for controlled shutdown
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, os.Kill)

	log.Printf("Listening on %v", srv.Addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	// Async shutdown allows ongoing requests to finish
	<-gracefulStop
	log.Println("Shutting down")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
	defer cancel()
	err := srv.Shutdown(ctx)

	if err != nil {
		log.Printf("error shutting down: %v", err)
	}
}
