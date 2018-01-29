# Translate Server
## What?
* A server that bundles multiple translation services as a single HTTP API.
* A server that caches repeated requests for the same translation.
* TOOD A server that pre-loads translations that may be requested in the future.

## Really?
* This is work in progress.
* Google Cloud Translation already works. Set `GOOGLE_API_KEY` environment variable to your API key.
* Caching works too, it also persists to the file system, under `cache.gob` (`/go/src/translate-server/cache.gob` in the Docker image)

## How?
* Run the Docker: 
    
    `docker run -p 8080:8080 -e GOOGLE_API_KEY=$GOOGLE_API_KEY kuboschek/translate-server`
    
* Send HTTP `POST` requests to port `8080` of the machine or container this is running in.
    * Send the content to be translated as `text/plain` in the request body.
    * Use the `Content-Language` header to indicate what language to translate from.
    * Use the `Accept-Language` header to indicate what language to translate to.
* With status `200 OK`, you'll get a translation back in the response body.
* With any other status, check the response content to see what happened.

### Using `cURL`
    curl -X POST \
    'http://localhost:8080/' \
    -H 'accept-language: en' \
    -H 'content-language: de' \
    -d 'Also wirklich!'

## Design Decisions
This section is just ideas - will be expanded in the future.
### Language
* Go is fast
* Go is safe
* Go has libraries

### Cache Structure
* Two-stage map
    * Simple to use in Go
    * Internally it's a hashmap, and thus fast

* TODO Persist with JSON
    * Allows for pre-filling of manually curated translations
    * Allows for cache sharing between multiple implementations

### Translation Backends
* Share memory by communicating -> channels used for response
    * Allows for pre-caching with Goroutines stashing to cache
* One interface -> remains pluggable

### Request Handler
* Doesn't care which backends it uses
* Fails over from one to the next
* TODO Moves failing backends to end of list

### Client Interface
* Uses standard HTTP headers -> compatible with many things already
* Uses no transfer encoding like JSON -> text is not structured data