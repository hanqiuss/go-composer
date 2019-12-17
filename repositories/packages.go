package repositories

import (
	"fmt"
	"go-composer/util"
	"sync"
)

var depend = make(map[string]*Project)
var repoList map[string]Repository
var blackList = map[string]bool{
	"facebook/php-webdriver": true,
}
var dependLock sync.Mutex
var depMap sync.Map

func GetDep(jsonPackage *JsonPackage) map[string]*Project {
	_, ok := depMap.LoadOrStore(jsonPackage.Name, true)
	if ok {
		return nil
	}
	if repoList == nil {
		repoList = CreateManager(jsonPackage)
	}

	ch := make(chan int)
	count := 0
	for name := range jsonPackage.Require {
		//jsonPackage.Require[name] = ver
		if !FilterRequire(&name) {
			continue
		}
		count++
		go func(name string) {
			defer metaDataGettingList.Delete(name)
			has := false
			for _, repo := range repoList {
				if repo.Has(name) {
					has = true
					ret := repo.GetPackages(name)
					if ret != nil {
						metaDataReadyList.Store(name, true)
						dependLock.Lock()
						depend[name] = ret
						dependLock.Unlock()
						p := JsonPackage{
							Name:    name,
							Require: make(map[string]string),
						}
						for _, v := range ret.Packages {
							for kk, vv := range v.Package.Require {
								p.Require[kk] = vv
							}
						}
						GetDep(&p)
					} else {
						failedList.Store(name, true)
						fmt.Println("get package failed", name, repo)
					}
				}
			}
			if !has {
				fmt.Println("lose package : ", name)
			}
			ch <- 1
		}(name)
	}
	for count > 0 {
		count--
		<-ch
	}
	return depend
}

var metaDataReadyList sync.Map
var metaDataGettingList sync.Map
var failedList sync.Map

func FilterRequire(name *string) bool {
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
	if !util.FilterRequire(*name) {
		return false
	}
	if _, ok = blackList[*name]; ok {
		return false
	}
	_, ok = metaDataGettingList.LoadOrStore(*name, true)
	if ok {
		return false
	}
	return true
}
