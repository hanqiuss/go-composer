package repositories

import (
	"composer/cache"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Composer struct {
	Url          string
	ProvidersUrl string
	Cache        cache.Base
}

func NewComposer(url string) *Composer {
	if url == "" {
		url = "https://packagist.jp/p/"
	}
	return &Composer{
		Url:          url,
		ProvidersUrl: "",
		Cache:        cache.Base{},
	}
}

func (c *Composer) GetPackages(name string) *JsonVersionPackages {
	url := c.getRepoUrl(name)
	data := c.Cache.GetManifest(name, url)
	var list JsonPackages
	parse := false
	if len(data) > 0 {
		err := json.Unmarshal(data, &list)
		if err == nil {
			parse = true
		}
	}
	if !parse {
		data := c.getManifestRemote(url)
		if len(data) == 0 || json.Unmarshal(data, &list) != nil {
			return nil
		}
		c.Cache.WriteManifest(name, data)
	}
	packages, ok := list.Packages[name]
	if !ok {
		return nil
	}
	return packages
}

func (c *Composer) getManifestRemote(url string) (r []byte) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return body
}

func (c *Composer) getRepoUrl(name string) string {
	return c.Url + name + ".json"
}
