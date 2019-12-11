package repositories

import (
	"fmt"
	"github.com/Masterminds/semver"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type JsonPackage struct {
	Name        string            `json:"name"`
	Description string            `json:"name"`
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string `json:"require"`
	Version     string            `json:"version"`
	Dist        struct {
		Type      string `json:"type"`
		Url       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"dist"`
	Source struct {
		Type      string `json:"type"`
		Url       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source"`
	Type     string                       `json:"type"`
	AutoLoad map[string]map[string]string `json:"autoload"`
	License  string                       `json:"license"`
}
type JsonVersionPackages map[string]*JsonPackage
type JsonPackages struct {
	Packages map[string]*JsonVersionPackages // [name:Packages]
}
type JsonLock struct {
	Packages []JsonPackage
}
type Package struct {
	Version *semver.Version
	Package *JsonPackage
}
type Packages []*Package
type Project struct {
	Constraints map[string]*semver.Constraints
	Packages    Packages
	Repository  *Composer
}

func (p Packages) Len() int           { return len(p) }
func (p Packages) Less(i, j int) bool { return p[i].Version.LessThan(p[j].Version) }
func (p Packages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

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
	ch := make(chan int)
	count := 0
	for name, ver := range jsonPackage.Require {
		ver = ReWriteVersion(ver)
		jsonPackage.Require[name] = ver
		if !filterRequire(&name, &ver) {
			continue
		}
		count++
		go func(name string) {
			defer metaDataGettingList.Delete(name)
			ret := repo.GetPackages(name)
			if ret != nil {
				metaDataReadyList.Store(name, true)
				depend[name] = ret
				dep := ret.Packages[0].Package
				GetDep(dep)
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
