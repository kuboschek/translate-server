[![Travis](https://img.shields.io/travis/kuboschek/translate-server.svg)](https://travis-ci.org/kuboschek/translate-server)
[![Docker Build Status](https://img.shields.io/docker/build/kuboschek/translate-server.svg)](https://hub.docker.com/r/kuboschek/translate-server/)
# Translate Server

## The Problem

In a service heavily relying on user generated content, scaling globally poses the challenge of seeding content for a
new region, to initially bring a platform to self-sufficiency. One common growth hack to circumvent this issue is
using machine translations to display content only available in other langauges in a user's native languages.

This can require using local APIs in markets with restricted access to mainstream services like Google Cloud Translation
or Microsoft's Cognitive Platform. Therefore, I propose a service which abstracts other translation APIs behind a common
HTTP API. The backend implementation uses Golang and allows for failover between any number of services configured. New
upstream service connectors are implemented with one method. Additionally, this service optionally connects to a cache
backend to store translations results. This accomodates upstream services that impose restrictive rate or total
translation limits.

## Scope

Currently, there is only a backend implementation. Furthermore, there is currently no backend for external cache services,
and no expiry timer for cached items.

Future improvements planned are:
 * Integration tests for upstreams
 * Concurrently calling multiple upstreams
 * Implement an authorization mechanism for clients
 * Enabling load-balancing to multiple upstreams / expanding failover options
 * Controlling the cache backend / settings at runtime
 * Backfilling / cache warming for likely future translations
 * Implementing a streaming RPC endpoint
 * Front-end to manually perform translations
 * API for injecting company-specific vocabulary
 * Allowing for templated translations / placeholders
 * Adding support for more translation services, and cache backends

## Technical Choices

As a language, Go was a good fit, because it provides both concurrency primitives and a well-tested HTTP implementation.
An added benefit is the speedup over Python / Java. This is a simple, low-level abstraction, so the faster the better.
For the external interface, I decided to implement a single HTTP endpoint, and used `Content-Langauge` and
`Accept-Language` headers. There is a spec, so this is easier to support.

The caching backend is implemented as a simple `Put` / `Get` / `Has` interface, allowing for straightforward expansion
to use external key-value stores. Translation services are called asynchronously - this enables future changes toward
calling multiple upstream services concurrently, and selecting the first one to answer. Currently one service is called
at a time. Failover is currently implemented by pushing a service to the end of the queue for requests to run on.

Upstream credentials are passed as environment variables. Currently, there are two variables processed:
 * `GOOGLE_API_KEY`: If specified, enable the Google Cloud Translation backend with given key.
 * `ENABLE_MOCK`: If specified, enables the mock backend. This is used for testing upstream failure handling.

### Testing Strategy

* The cache package is fully unit tested. It plays a part in every request and is critical to reducing upstream load.
* `handler.go` has full coverage as well. It handles every request, making it a critical piece of code.
* `sanity_test.go` contains tests running `gofmt` and `govet`. Code is more often read than written, so this was a
no-brainer.

## About The Author
My name is Leonhard Kuboschek. This code uses libraries written by many people. Thank you. All code contained within
this repository was authored solely by me. I'm currently working full-time at a startup (50 hrs / week),
and we have a launch coming up. Thus I've been working on this after hours, and on the weekends only.
Additionally, this is one of my earlier projects utilizing Go as a language.

 * [Personal Homepage](http://kuboschek.me)

## HTTP Interface

The server accepts HTTP `POST` requests to port 8080, the path is `/`.

The headers required are:
 * `Content-Language` specifying the language that the request content is assumed to be in.
 * `Accept-Language` specifying the target language.

The text to be translated is sent in the request body; The content type shall be `text/plain`.

The response will contain the `Content-Language` header for the target language, as well as the translated text in the
response body. In case of an error, standard HTTP status codes are used for signaling.

### Sample Deployment

There is a sample deployment running at translate dot leo dot codes. It's secured using HTTP Basic Auth to prevent
misuse. If you would like to receive testing credentials, please send me an email (see homepage links below).

### Sample request using `cURL`
    curl -X POST \
    'http://localhost:8080/' \
    -H 'accept-language: en' \
    -H 'content-language: de' \
    -d 'Also wirklich!'

## Docker Image
    `docker run -p 8080:8080 -e GOOGLE_API_KEY=$GOOGLE_API_KEY kuboschek/translate-server`
