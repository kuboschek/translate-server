// Package cache provides an interface and some implementations for caching translations
package cache

import "golang.org/x/text/language"

// Cache is the interface describing a translation cache.
type Cache interface {
	Put(sourcePhrase string, targetLang language.Tag, targetPhrase string) (err error)
	Has(sourcePhrase string, targetLang language.Tag) bool
	Get(sourcePhrase string, targetLang language.Tag) (targetPhrase string, err error)
}
