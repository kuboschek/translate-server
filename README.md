Design Decisions
    * Language
        * Go is fast
        * Go is safe
        * Go has libraries

    * Cache Structure
        * Two-stage map
            * Simple to use in Go
            * Internally it's a hashmap, and thus fast
        
        * TODO Persist with JSON
            * Allows for pre-filling of manually curated translations
            * Allows for cache sharing between multiple implementations

    * Translation Backends
        * Share memory by communicating -> channels used for response
            * Allows for pre-caching with Goroutines stashing to cache
    
    * Client Interface
        * Uses standard HTTP headers -> compatible with many things already
        * Uses no transfer encoding like JSON -> text is not structured data