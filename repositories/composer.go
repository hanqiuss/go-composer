package repositories

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"go-composer/cache"
	"go-composer/util"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Composer struct {
	Url          string
	ProvidersUrl string
	Cache        *cache.Base
	Providers    *JsonProvPack
	Host         string
}

func (c *Composer) Has(name string) bool {
	_, ok := c.Providers.Providers[name]
	return ok
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
		Providers:    &JsonProvPack{make(map[string]*JsonProvider)},
	}
	err = c.Init()
	if err != nil {
		fmt.Println(err)
	}
	return c
}
func (c *Composer) Init() error {
	urlStr := c.Url + "packages.json"
	body, err := util.DownloadAndSave(urlStr, filepath.Join(c.Cache.GetRepoDir(c.Host), "packages.json"), "")
	if len(body) == 0 {
		return err
	}
	repo := JsonRepos{}
	err = json.Unmarshal(body, &repo)
	if err != nil {
		return fmt.Errorf("json decode packages.json error : %s", err)
	}
	c.ProvidersUrl = repo.ProvidersUrl
	atom := new(sync.Mutex)
	wg := sync.WaitGroup{}
	for k, v := range repo.ProviderIncludes {
		wg.Add(1)
		go func(k string, v *JsonProvider) {
			urlStr := strings.Replace(c.Url+k, "%hash%", v.Hash, 1)
			file := strings.Replace(k[2:], "$%hash%", "", 1)
			file = filepath.Join(c.Cache.GetRepoDir(c.Host), "p-"+file)
			body, err = util.DownloadAndSave(urlStr, file, v.Hash)
			p := JsonProvPack{}
			err = json.Unmarshal(body, &p)
			if err != nil {
				fmt.Printf("json decode %s failed : %s \r\n", urlStr, err)
			}
			atom.Lock()
			for pkg, pro := range p.Providers {
				c.Providers.Providers[pkg] = pro
			}
			atom.Unlock()
			wg.Done()
		}(k, v)
	}
	wg.Wait()
	count := 0
	for range c.Providers.Providers {
		count++
	}
	fmt.Println("count : ", count, time.Now())
	return nil
}
func (c *Composer) GetPackages(name string) *Project {
	hash_ := c.Providers.Providers[name].Hash
	fUrl := c.getRepoUrl(name, hash_)
	data := c.Cache.GetManifest(name, fUrl, hash_)
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
		Packages:    getPackages(packages),
		Repository:  c,
	}
}

func (c *Composer) getRepoUrl(name, h string) string {
	r := strings.Replace(c.ProvidersUrl, "%package%", name, 1)
	return c.Url + strings.Replace(r, "%hash%", h, 1)
}
func getPackages(packages *JsonVersionPackages) Packages {
	ret := Packages{}
	for v, p := range *packages {
		version, err := semver.NewVersion(v)
		if err != nil || version == nil {
			continue
		}
		ret = append(ret, &Package{version, p})
	}
	sort.Sort(sort.Reverse(ret))
	return ret
}
