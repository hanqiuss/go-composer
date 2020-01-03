package repositories

import (
	"encoding/json"
	"fmt"
	"go-composer/cache"
	"net/url"
	"strings"
)

type Npm struct {
	Url   string
	Cache *cache.Base
	Host  string
}

func NewNpm(urls string) *Npm {
	if urls == "" {
		urls = "https://registry.npmjs.org/"
	}
	urls = strings.Trim(urls, "/") + "/"
	u, err := url.Parse(urls)
	if err != nil {
		fmt.Println("repository composer : url error ", err)
	}
	cacheBase := cache.NewCacheBase()
	cacheBase.CreateManifestDir(u.Host)
	return &Npm{
		Url:   urls,
		Cache: cacheBase,
		Host:  u.Host,
	}
}
func (r *Npm) Has(name string) bool {
	return IsNpmPkg(name)
}
func (r *Npm) GetPackages(name string) *Project {
	fUrl := r.getRepoUrl(name)
	data := r.Cache.GetManifest(name, fUrl)
	if len(data) == 0 {
		fmt.Println("get manifest failed", fUrl)
		return nil
	}
	var list JsonNpmPackages
	if err := json.Unmarshal(data, &list); err != nil {
		fmt.Println("json decode failed", err)
		return nil
	}
	var packages = make(JsonVersionPackages)
	for v, p := range list.Versions {
		p.Name = name
		packages[v] = JsonNpmToComposer(p)
	}
	return &Project{
		Constraints: make(map[string]bool),
		Conflicts:   map[string]bool{},
		Packages:    getPackages(&packages),
	}
}

func (r *Npm) getRepoUrl(name string) string {
	name = strings.Replace(name, "bower-asset/", "", 1)
	name = strings.Replace(name, "npm-asset/", "", 1)
	if strings.Contains(name, "--") {
		name = "@" + strings.Replace(name, "--", "/", 1)
	}
	return r.Url + name
}

func IsNpmPkg(name string) bool {
	return strings.Contains(name, "bower-asset") || strings.Contains(name, "npm-asset")
}
