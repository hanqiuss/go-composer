package cache

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Base struct {
	dir         string
	repoDir     string
	filesDir    string
	manifestPre string
}

var ComposerCache Base

func NewCacheBase() *Base {
	if ComposerCache.dir != "" {
		return &ComposerCache
	}
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use cwd path ")
		path, _ := os.Getwd()
		cacheDir = filepath.Join(path, "cache$$")
	}
	cacheDir = filepath.Join(cacheDir, "Composer")
	repoDir := filepath.Join(cacheDir, "repo", "https---repo.packagist.org")
	cacheFileDir := filepath.Join(cacheDir, "files")
	_, err := os.Stat(repoDir)
	if err != nil {
		err = os.MkdirAll(repoDir, os.ModePerm)
		if err != nil {
			fmt.Println("make cache dir failed")
			os.Exit(-1)
		}
	}
	_, err = os.Stat(cacheFileDir)
	if err != nil {
		err = os.MkdirAll(repoDir, os.ModePerm)
		if err != nil {
			fmt.Println("make cache dir failed")
			os.Exit(-1)
		}
	}
	ComposerCache = Base{
		dir:         cacheDir,
		repoDir:     repoDir,
		filesDir:    cacheFileDir,
		manifestPre: "provider-",
	}
	return &ComposerCache
}
func (c *Base) GetManifest(name, url string) (r []byte) {
	file := c.getManifestPath(name)
	info, err := os.Stat(file)
	if err == nil {
		resp, err := http.Head(url)
		if err != nil {
			fmt.Println("get url failed : ", url)
		} else {
			timeStr := resp.Header.Get("date")
			if timeStr == "" { //read from url
				return
			}
			time, err := http.ParseTime(timeStr)
			if err != nil {
				return nil
			}
			if time.After(info.ModTime()) {
				return
			}
		}
		ret, err := ioutil.ReadFile(file)
		if err != nil {
			return
		}
		return ret
	}
	return
}
func (c *Base) getManifestPath(name string) string {
	return filepath.Join(c.repoDir, c.manifestPre+strings.ReplaceAll(name, "/", "$")) + ".json"
}
func (c *Base) WriteManifest(name string, body []byte) bool {
	file := c.getManifestPath(name)
	err := ioutil.WriteFile(file, body, os.ModePerm)
	return err == nil
}
