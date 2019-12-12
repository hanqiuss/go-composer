package repositories

import (
	"fmt"
	"github.com/Masterminds/semver"
	"regexp"
	"sort"
	"strings"
	"sync"
)

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

var repo = NewComposer("")
var depend = make(map[string]*Project)

func GetDep(jsonPackage *JsonPackage) map[string]*Project {
	//err := repo.Init()
	//fmt.Println(err)

	ch := make(chan int)
	count := 0
	for name, ver := range jsonPackage.Require {
		ver = ReWriteVersion(ver)
		jsonPackage.Require[name] = ver
		if !FilterRequire(&name, &ver) {
			continue
		}
		count++
		go func(name string) {
			defer metaDataGettingList.Delete(name)
			ret := repo.GetPackages(name)
			if ret != nil {
				metaDataReadyList.Store(name, true)
				depend[name] = ret
				for _, v := range ret.Packages {
					GetDep(v.Package)
				}
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
func ReWriteVersion(v string) string {
	v = strings.ReplaceAll(strings.ReplaceAll(v, "||", "|"), "|", "||")
	return strings.ReplaceAll(v, "@", "-")
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
		fmt.Printf("require version %s error : %s\r\n", *ver, err)
		return false
	}
	_, ok = metaDataGettingList.LoadOrStore(*name, true)
	if ok {
		return false
	}
	return true
}
