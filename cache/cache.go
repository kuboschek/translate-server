// Package cache provides an interface and some implementations for caching translations
package cache

import "golang.org/x/text/language"

// Cache is the interface describing a translation cache.
type Cache interface {
	// Put makes a translation available for subsequent calls to other methods of this cache
	Put(sourcePhrase string, targetLang language.Tag, targetPhrase string) (err error)

	// Has returns true if there is a translation matching the parameters
	Has(sourcePhrase string, targetLang language.Tag) bool

	// Get returns the translation matching the parameters, or an error, if it could not be retrieved,
	// or is not present in cache.
	Get(sourcePhrase string, targetLang language.Tag) (targetPhrase string, err error)
}
