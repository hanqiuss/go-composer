package repositories

import "go-composer/cache"

type Npm struct {
	Url          string
	ProvidersUrl string
	Cache        *cache.Base
	Host         string
}
