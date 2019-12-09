package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type Package struct {
	MetaData
	Version string
	Dist    struct {
		Type string
		Url  string
	}
	Source struct {
		Type string
		Url  string
	}
}

type Packages map[string]*Package // [version:Package]

type PackageManifest struct {
	Packages map[string]*Packages // [name:Packages]
}

var cacheDir = ""
var repo = "https---repo.packagist.org"
var repoUrl = "https://packagist.jp/p/"
var repoDir = ""
var cacheFileDir = ""
var metaDataReadyList sync.Map
var metaDataGettingList sync.Map
var failedList sync.Map
var versionMap map[string]*VersionPack

func init() {
	cacheDir, _ = os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use cwd path ")
		path, _ := os.Getwd()
		cacheDir = filepath.Join(path, "cache$$")
	}
	cacheDir = filepath.Join(cacheDir, "Composer")
	repoDir = filepath.Join(cacheDir, "repo", repo)
	cacheFileDir = filepath.Join(cacheDir, "files")
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
	versionMap = make(map[string]*VersionPack)
}

func getPackages(list map[string]string) {
	ch := make(chan int)
	count := 0
	for name, ver := range list {
		fmt.Println(name, ver)
		ver = strings.ReplaceAll(strings.ReplaceAll(ver, "||", "|"), "|", "||")
		ver = strings.ReplaceAll(ver, "@", "-")
		if !filterRequire(&name, &ver) {
			continue
		}
		count++
		go func(name, ver string, ch chan int) {
			defer metaDataGettingList.Delete(name)
			_getPackage(&name, &ver)
			ch <- 1
		}(name, ver, ch)
	}
	for count > 0 {
		count--
		<-ch
	}

}

func filterRequire(name, ver *string) bool {
	_, ok := metaDataGettingList.Load(*name)
	if ok {
		return false
	}
	_, ok = failedList.Load(*name)
	if ok {
		return false
	}
	_, ok = metaDataReadyList.Load(*name)
	if ok {
		return false
	}
	ok, _ = regexp.MatchString("^php$", *name)
	if ok {
		return false
	}
	ok, _ = regexp.MatchString("^(ext|lib)-.*", *name)
	if ok {
		return false
	}
	_, err := semver.NewConstraint(*ver)
	if err != nil {
		fmt.Printf("require version %s error : %s\r\n", *ver, err)
		return false
	}
	_, ok = metaDataGettingList.LoadOrStore(*name, true)
	if ok {
		return false
	}
	return true
}

func _getPackage(name, ver *string) {
	file := filepath.Join(repoDir, "provider-"+strings.ReplaceAll(*name, "/", "$")) + ".json"
	fmt.Println(file)
	url := getRepoUrl(name)
	data := cacheGetManifest(&file, &url)
	parse := false
	var list PackageManifest
	if len(data) > 0 {
		err := json.Unmarshal(data, &list)
		if err == nil {
			parse = true
		}
	}
	if !parse {
		data := getManifest(name, &file)
		if len(data) == 0 || json.Unmarshal(data, &list) != nil {
			failedList.Store(*name, true)
			return
		}
	}
	packages, ok := list.Packages[*name]
	if !ok {
		fmt.Println("the manifest have not the packages : ", *name)
		failedList.Store(*name, true)
		return
	}
	metaDataReadyList.Store(*name, true)
	fmt.Println(*name)
	versionPack := FilterVersion(*packages, *ver)
	if versionPack == nil {
		fmt.Println("can't get useful version", *name, ver)
		return
	}
	versionMap[*name] = versionPack
	if versionPack.Package[0].p.MetaData.Require != nil {
		getPackages(versionPack.Package[0].p.MetaData.Require)
	}
}

func getManifest(name *string, fileName *string) (ret []byte) {
	url := getRepoUrl(name)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("get package %s error %s , ", *name, err)
		return
	}
	if resp.StatusCode != 200 {
		fmt.Printf("get package %s metadata status code %d url \r\n%s , ", *name, resp.StatusCode, url)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("get package %s metadata error, read body error %s url \r\n%s , ", *name, err, url)
		return
	}
	err = ioutil.WriteFile(*fileName, body, os.ModePerm)
	if err != nil {
		fmt.Printf("write cache file error \r\n file : %s", *fileName)
		return
	}
	return body
}

func getRepoUrl(name *string) string {
	return repoUrl + *name + ".json"
}

func cacheGetManifest(file, url *string) (ret []byte) {
	info, err := os.Stat(*file)
	if err == nil {
		resp, err := http.Head(*url)
		if err != nil {
			fmt.Println("get url failed : ", *url)
		} else {
			timeStr := resp.Header.Get("date")
			if timeStr == "" { //read from url
				return
			}
			time, err := http.ParseTime(timeStr)
			if err != nil {
				return
			}
			if time.After(info.ModTime()) {
				return
			}
		}
		ret, err := ioutil.ReadFile(*file)
		if err != nil {
			return []byte{}
		}
		return ret
	}
	return
}
