package solver

import (
	"fmt"
	"github.com/Masterminds/semver"
	"go-composer/repositories"
	"sort"
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
	setDep(dep)
	for k, v := range dep {
		fmt.Println(k, *v, v.Constraints)
	}

}

var sel = make(map[string]int)

func setDep(dep map[string]*repositories.Project) {
	for root, project := range dep {
		sel[root] = 0
		for depName, v := range (*project.Packages)[0].Package.Require {
			depPac, ok := dep[depName]
			if !ok {
				fmt.Println("package lost", depName)
				continue
			}
			constr, err := semver.NewConstraint(v)
			if err != nil {
				continue
			}
			depPac.Constraints[root] = constr
		}
	}
}

func solveDep(dep map[string]*repositories.Project) {

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
	sort.Sort(sort.Reverse(depSort))
	return depSort
}
