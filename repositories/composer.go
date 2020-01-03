package repositories

import (
	"encoding/json"
	"fmt"
	"go-composer/cache"
	"net/url"
	"sort"
	"strings"
)

type Composer struct {
	Url          string
	ProvidersUrl string
	Cache        *cache.Base
	Host         string
}

func (c *Composer) Has(name string) bool {
	return !IsNpmPkg(name)
}

func NewComposer(urls string) *Composer {
	if urls == "" {
		urls = "https://packagist.jp"
	}
	urls = strings.Trim(urls, "/") + "/"
	u, err := url.Parse(urls)
	if err != nil {
		fmt.Println("repository composer : url error ", err)
	}
	cacheBase := cache.NewCacheBase()
	cacheBase.CreateManifestDir(u.Host)
	c := &Composer{
		Url:          urls,
		ProvidersUrl: "",
		Cache:        cacheBase,
		Host:         u.Host,
	}
	return c
}
func (c *Composer) GetPackages(name string) *Project {
	fUrl := c.getRepoUrl(name)
	data := c.Cache.GetManifest(name, fUrl)
	if len(data) == 0 {
		fmt.Println("get manifest failed", fUrl)
		return nil
	}
	var list JsonPackages
	if err := json.Unmarshal(data, &list); err != nil {
		fmt.Println("json decode failed", err)
		return nil
	}
	packages, ok := list.Packages[name]
	if !ok {
		return nil
	}

	return &Project{
		Constraints: make(map[string]bool),
		Conflicts:   map[string]bool{},
		Packages:    getPackages(packages),
	}
}

func (c *Composer) getRepoUrl(name string) string {
	return c.Url + "p/" + name + ".json"
}
func getPackages(packages *JsonVersionPackages) Packages {
	ret := Packages{}
	for v, p := range *packages {
		pkg := NewPackage(v, p)
		if pkg == nil {
			continue
		}
		ret = append(ret, pkg)
	}
	sort.Sort(sort.Reverse(ret))
	return ret
}
