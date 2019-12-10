package solver

import (
	"fmt"
	"github.com/Masterminds/semver"
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
		Packages:    &repositories.Packages{&repositories.Package{Version: rootVersion, Package: p}},
		Repository:  nil,
	}
	dependList = dep
	fmt.Println("start solve sat", time.Now())
	setDep()
	solveDep()
	solveDep()
	fmt.Println("11111")
}

var sel = make(map[string]int)
var dependList map[string]*repositories.Project

func setDep() {
	for root, project := range dependList {
		sel[root] = 0
		for depName, v := range (*project.Packages)[0].Package.Require {
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

func solveDep() {
	sortDep := getSort(dependList)
	for _, count := range sortDep {
		if count.name == "root" {
			continue
		}
		fmt.Println(*count)
		ret := solveDepByName(count.name)
		if !ret {
			fmt.Println("solve error")
			return
		}
	}
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
			if sel[depByName] >= len(*dependList[depByName].Packages) {
				return false
			}
			str := (*dependList[depByName].Packages)[sel[depByName]].Package.Require[name]
			curCts, err := semver.NewConstraint(str)
			if err != nil {
				sel[depByName]++
				fmt.Println("error constraints ", str)
				goto begin
			}
			index := getCheckVersionIndex(curCts, *dependList[depByName].Packages)
			if index == len(*dependList[depByName].Packages) {
				fmt.Printf("package %s need %s %s, no match", depByName, name, str)
				return false
			}

			if min > index {
				min = index
				minName = depByName
			}
		}
		// check is some version match all constraints
		for ; min < len(*dependList[name].Packages); min++ {
			p := (*dependList[name].Packages)[min]
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
			for _, v := range *dependList[name].Packages {
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
