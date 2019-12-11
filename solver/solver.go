package solver

import (
	"fmt"
	"github.com/Masterminds/semver"
	"go-composer/cache"
	"go-composer/repositories"
	"sort"
	"strings"
	"time"
)

func Solver(p *repositories.JsonPackage) {
	if p.Require == nil {
		p.Require = make(map[string]string)
	}
	for name, v := range p.RequireDev {
		p.Require[name] = v
	}
	rootVersion, err := semver.NewVersion(p.Version)
	if err != nil {
		rootVersion = nil
	}
	dep := repositories.GetDep(p)
	dep["root"] = &repositories.Project{
		Constraints: make(map[string]*semver.Constraints),
		Packages:    repositories.Packages{&repositories.Package{Version: rootVersion, Package: p}},
		Repository:  nil,
	}
	dependList = dep
	fmt.Println("start solve sat", time.Now())
	setDep()
	for solveDep() && !checkDep() {
	}
	install := getInstallList("root")
	fmt.Println("check : ", checkDep())
	count := 0
	ch := make(chan int)
	for _, v := range install {
		if v.Dist.Type != "zip" {
			fmt.Println("dist type error", v)
		} else {
			count++
			go func(p *repositories.JsonPackage, ch chan int) {
				cacheObj := cache.NewCacheBase()
				cacheObj.CacheFiles(p.Name, p.Dist.Url, p.Dist.Type)
				err := cacheObj.Install(p.Name, p.Dist.Url, p.Dist.Type)
				if err != nil {
					fmt.Println(err)
				}
				ch <- 1
			}(v, ch)
		}
	}
	for count > 0 {
		count--
		<-ch
	}
	fmt.Println("11111")
}

var sel = make(map[string]int)
var dependList map[string]*repositories.Project

/* set dependence link */
func setDep() {
	for root, project := range dependList {
		sel[root] = 0
		for depName, v := range project.Packages[0].Package.Require {
			depPac, ok := dependList[depName]
			if !ok {
				if strings.Contains(depName, "/") {
					fmt.Println("package lost", depName)
				}
				continue
			}
			constr, err := semver.NewConstraint(v)
			if err != nil {
				fmt.Println("error constraints", depName, v)
				continue
			}
			depPac.Constraints[root] = constr
		}
	}
}
func solveDep() bool {
	sortDep := getSort(dependList)
	for _, count := range sortDep {
		if count.name == "root" {
			continue
		}
		ret := solveDepByName(count.name)
		if !ret {
			fmt.Println("solve error")
			return false
		}
	}
	return true
}
func solveDepByName(name string) bool {
	cts := dependList[name].Constraints
	for {
	begin:
		min := 10000000
		minName := ""
		ctsList := make([]*semver.Constraints, 0)
		// get the top index of  the match version
		for depByName := range cts {
			if sel[depByName] >= len(dependList[depByName].Packages) {
				return false
			}
			str, ok := dependList[depByName].Packages[sel[depByName]].Package.Require[name]
			if !ok {
				continue
			}
			str = repositories.ReWriteVersion(str)
			curCts, err := semver.NewConstraint(str)
			if err != nil {
				sel[depByName]++
				fmt.Println("error constraints ", str)
				goto begin
			}
			ps := dependList[name].Packages
			index := getCheckVersionIndex(curCts, ps)
			if index == len(dependList[depByName].Packages) {
				fmt.Printf("package %s need %s %s, no match", depByName, name, str)
				return false
			}
			ctsList = append(ctsList, curCts)
			if min > index {
				min = index
				minName = depByName
			}
		}
		if len(ctsList) == 0 {
			return true
		}
		n := len(dependList[name].Packages)
		// check is some version match all constraints
		for min < n {
			p := (dependList[name].Packages)[min]
			check := true
			for _, cts := range ctsList {
				if !cts.Check(p.Version) {
					check = false
					break
				}
			}
			if check {
				sel[name] = min
				return true
			}
			min++
		}
		// downgrade the dependBy version which has the max require version(constraints)
		_, ok := sel[minName]
		if ok {
			sel[minName]++
		} else {
			fmt.Println("no version match, require ", name)
			for k, v := range cts {
				fmt.Println(k, v)
			}
			for _, v := range dependList[name].Packages {
				fmt.Println(name, v.Version)
			}
			return false
		}
	}
}
func getCheckVersionIndex(cts *semver.Constraints, p repositories.Packages) int {
	i := len(p)
	for j := 0; j < i; j++ {
		if p[j].Version == nil {
			continue
		}
		if cts.Check(p[j].Version) {
			i = j
		}
	}
	if i == len(p) {
		fmt.Println("no useful version ", cts, p[0].Package.Name)
	}
	return i
}
func checkDep() bool {
	for name, project := range dependList {
		if name == "root" {
			continue
		}
		ver := dependList[name].Packages[sel[name]].Version
		for depByName := range project.Constraints {
			str, ok := dependList[depByName].Packages[sel[depByName]].Package.Require[name]
			if !ok {
				continue
			}
			str = repositories.ReWriteVersion(str)
			ct, err := semver.NewConstraint(str)
			if err != nil {
				fmt.Println("check error : error constraints", str)
			}
			if !ct.Check(ver) {
				return false
			}
		}

	}
	return true
}
func getInstallList(root string) (list map[string]*repositories.JsonPackage) {
	list = make(map[string]*repositories.JsonPackage)
	project, ok := dependList[root]
	if !ok {
		return
	}
	for name := range project.Packages[sel[root]].Package.Require {
		p, ok := dependList[name]
		if !ok {
			continue
		}
		list[name] = p.Packages[sel[name]].Package
		for k, v := range getInstallList(name) {
			list[k] = v
		}
	}
	return list
}

type depS struct {
	name  string
	count int
}
type depSs []*depS

func (d depSs) Len() int           { return len(d) }
func (d depSs) Less(i, j int) bool { return d[i].count < d[j].count }
func (d depSs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

/*
* 按依赖排序
 */
func getSort(dep map[string]*repositories.Project) depSs {
	depSort := depSs{}
	for name, project := range dep {
		count := 0
		for range project.Constraints {
			count++
		}
		depSort = append(depSort, &depS{
			name:  name,
			count: count,
		})
	}
	sort.Sort(depSort)
	return depSort
}
