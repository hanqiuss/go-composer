package repositories

import (
	"encoding/json"
	"fmt"
	"go-composer/cache"
	"go-composer/util"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Composer struct {
	Url          string
	ProvidersUrl string
	Cache        *cache.Base
	Providers    JsonProvPack
}

func NewComposer(url string) *Composer {
	if url == "" {
		url = "https://packagist.jp/"
	}
	return &Composer{
		Url:          url,
		ProvidersUrl: "",
		Cache:        cache.NewCacheBase(),
	}
}
func (c Composer) Init() error {
	url := c.Url + "packages.json"
	body, err := util.DownloadAndSave(url, filepath.Join(c.Cache.GetRepoDir(), "packages.json"), "")
	if len(body) == 0 {
		return err
	}
	repo := JsonRepos{}
	err = json.Unmarshal(body, &repo)
	if err != nil {
		return fmt.Errorf("json decode packages.json error : %s", err)
	}
	atom := new(sync.Mutex)
	ch := make(chan int)
	for k, v := range repo.ProviderIncludes {
		go func(k string, v *JsonProvider, ch chan int) {
			defer func() { ch <- 1 }()
			fmt.Println(k, v)
			url := strings.Replace(c.Url+k, "%hash%", v.Hash, 1)
			file := strings.Replace(k[2:], "$%hash%", "", 1)
			file = filepath.Join(c.Cache.GetRepoDir(), file)
			body, err = util.DownloadAndSave(url, file, v.Hash)
			atom.Lock()
			err = json.Unmarshal(body, &c.Providers)
			atom.Unlock()
			if err != nil {
				fmt.Printf("json decode %s failed : %s \r\n", url, err)
			}
		}(k, v, ch)
	}
	for range repo.ProviderIncludes {
		<-ch
	}
	count := 0
	for range c.Providers.Providers {
		count++
	}
	fmt.Println("count : ", count, time.Now())
	return nil
}
func (c *Composer) GetPackages(name string) *Project {
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
		if len(data) == 0 {
			fmt.Println("get manifest failed", url)
			return nil
		}
		if err := json.Unmarshal(data, &list); err != nil {
			fmt.Println("json decode failed", err)
			return nil
		}
		c.Cache.CacheManifest(name, data)
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

func (c *Composer) getManifestRemote(url string) (r []byte) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Println("http error ", err)
		return
	}
	defer util.Close(resp.Body)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read http body error ", err)
		return
	}
	return body
}

func (c *Composer) getRepoUrl(name string) string {
	return c.Url + "p/" + name + ".json"
}
