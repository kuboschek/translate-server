package cache

import (
	"github.com/pkg/errors"
	"golang.org/x/text/language"
	"sync"
)

type memoryCache map[string]map[language.Tag]string

var (
	// Memory is an instance of an in-memory cache.
	memoryLock     sync.RWMutex
	Memory         memoryCache
	phraseNotFound = errors.New("phrase not found in memory cache")
)

func init() {
	Memory = make(memoryCache)
}

func (p memoryCache) Put(sourcePhrase string, targetLang language.Tag, targetPhrase string) error {
	memoryLock.Lock()
	defer memoryLock.Unlock()

	if p[sourcePhrase] == nil {
		p[sourcePhrase] = make(map[language.Tag]string)
	}

	p[sourcePhrase][targetLang] = targetPhrase

	return nil
}

func (p memoryCache) Has(sourcePhrase string, targetLang language.Tag) bool {
	memoryLock.RLock()
	defer memoryLock.RUnlock()

	phrases, hasPhrase := p[sourcePhrase]
	_, hasTargetLang := phrases[targetLang]
	return hasPhrase && hasTargetLang
}

func (p memoryCache) Get(sourcePhrase string, targetLang language.Tag) (targetPhrase string, err error) {
	memoryLock.RLock()
	defer memoryLock.RUnlock()

	phrases, ok := p[sourcePhrase]
	if !ok {
		return "", phraseNotFound
	}

	phrase, ok := phrases[targetLang]
	if !ok {
		return "", phraseNotFound
	}

	return phrase, nil
}
