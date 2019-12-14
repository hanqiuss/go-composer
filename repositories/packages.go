package repositories

import (
	"fmt"
	"github.com/Masterminds/semver"
	"go-composer/util"
	"regexp"
	"strings"
	"sync"
)

var depend = make(map[string]*Project)
var repoList map[string]Repository

func GetDep(jsonPackage *JsonPackage) map[string]*Project {
	if repoList == nil {
		repoList = CreateManager(jsonPackage)
	}
	ch := make(chan int)
	count := 0
	for name, ver := range jsonPackage.Require {
		ver = util.ReWriteVersion(ver)
		jsonPackage.Require[name] = ver
		if !FilterRequire(&name, &ver) {
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
						depend[name] = ret
						for _, v := range ret.Packages {
							GetDep(v.Package)
						}
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

func FilterRequire(name, ver *string) bool {
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
	if !strings.Contains(*name, "/") {
		return false
	}
	_, err := semver.NewConstraint(*ver)
	if err != nil {
		//fmt.Printf("require version%s %s error : %s\r\n",*name, *ver, err)
		return false
	}
	_, ok = metaDataGettingList.LoadOrStore(*name, true)
	if ok {
		return false
	}
	return true
}
