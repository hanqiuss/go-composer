package solver

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go-composer/cache"
	"go-composer/repositories"
	"go-composer/semver"
	"go-composer/template"
	"go-composer/util"
	"io/ioutil"
	"math"
	"runtime"
	"sort"
	"time"
)

func Solver(p *repositories.JsonPackage) {
	if p.Require == nil {
		p.Require = make(map[string]string)
	}
	if util.Conf.Dev {
		for name, v := range p.RequireDev {
			p.Require[name] = v
		}
	}
	rootVersion, err := semver.NewVersion(p.Version)
	if err != nil {
		rootVersion = nil
	}
	dep := repositories.GetDep(p)
	dep["root"] = &repositories.Project{
		Constraints: make(map[string]bool),
		Packages:    repositories.Packages{&repositories.Package{Version: rootVersion, Package: p}},
	}
	dependList = dep
	fmt.Println("start solve sat", time.Now())

	installList = getInstallList("root")
	setDep()
	for !checkDep() {
		if !solveDep() {
			fmt.Println("solver error")
			return
		}
		installList = getInstallList("root")
		setDep()
	}
	install()
	fmt.Println("end install sat", time.Now())
}

var sel = make(map[string]int)
var dependList map[string]*repositories.Project
var installList map[string]*repositories.JsonPackage

/* set dependence link */
func setDep() {
	for _, p := range dependList {
		p.Constraints = make(map[string]bool)
	}
	installList["root"] = dependList["root"].Packages[0].Package
	for root, p := range installList {
		for depName, ver := range p.Require {
			_, ok := dependList[depName]
			if !ok {
				continue
			}
			ver = util.ReWriteConstraint(ver)
			_, err := semver.NewConstraint(ver)
			if err != nil {
				fmt.Println("error constraints", depName, ver)
				continue
			}
			dependList[depName].Constraints[root] = true
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
			fmt.Println("solve error", count.name)
			return false
		}
	}
	return true
}
func solveDepByName(name string) bool {
	cts := dependList[name].Constraints
	for {
	begin:
		min := math.MaxInt64
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
			str = util.ReWriteConstraint(str)
			curCts, err := semver.NewConstraint(str)
			if err != nil {
				sel[depByName]++
				fmt.Println("error constraints ", str)
				goto begin
			}
			ps := dependList[name].Packages
			index := getCheckVersionIndex(curCts, ps)
			if index == len(dependList[name].Packages) {
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
			p := dependList[name].Packages[min]
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
		sel[minName]++
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
	for name := range installList {
		if name == "root" {
			continue
		}
		if sel[name] >= len(dependList[name].Packages) {
			return false
		}
		ver := dependList[name].Packages[sel[name]].Version
		for depByName := range dependList[name].Constraints {
			str, ok := dependList[depByName].Packages[sel[depByName]].Package.Require[name]
			if !ok {
				continue
			}
			str = util.ReWriteConstraint(str)
			ct, err := semver.NewConstraint(str)
			if err != nil {
				fmt.Println("check error : error constraints", str)
				continue
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
func install() {
	markDev()
	count := 0
	lock := repositories.JsonLock{}
	delete(installList, "root")
	numCpu := runtime.NumCPU() * 2
	pkgCh := make(chan *repositories.JsonPackage, numCpu)
	resultCh := make(chan int, numCpu)
	// run 4 concurrent parserWorkers
	for i := 0; i < numCpu; i++ {
		go installWorker(pkgCh, resultCh)
	}
	for k, v := range installList {
		if v.Dist.Type != "zip" && v.Dist.Type != util.NpmPkgType {
			fmt.Println("dist type error", v)
		} else {
			count++
			if dependList[k].IsDev {
				lock.PackagesDev = append(lock.PackagesDev, v)
			} else {
				lock.Packages = append(lock.Packages, v)
			}
			go func(v *repositories.JsonPackage) { pkgCh <- v }(v)
		}
	}
	// generated .lock file
	cData, _ := ioutil.ReadFile("composer.json")
	h := md5.Sum(cData)
	lock.Hash = hex.EncodeToString(h[:])

	for count > 0 {
		count--
		<-resultCh
	}
	close(pkgCh)
	close(resultCh)
	installEnd(lock)
}
func installEnd(lock repositories.JsonLock) {
	lock.Sort()
	err := util.JsonDataToFile("composer.lock", lock)
	if err != nil {
		fmt.Println("write lock file error ", err)
	}
	lock.Packages = append(lock.Packages, lock.PackagesDev...)
	lock.Sort()
	err = util.JsonDataToFile("vendor/composer/installed.json", lock.Packages)
	if err != nil {
		fmt.Println("write installed.json file error ", err)
	}
	err = template.Generated(installList)
	if err != nil {
		fmt.Println(err)
	}
}
func markDev() {
	if !util.Conf.Dev {
		return
	}
	root := dependList["root"].Packages[0].Package
	for k := range root.RequireDev {
		_markDev(k, true)
	}
	for k := range root.Require {
		_, ok := root.RequireDev[k]
		if !ok {
			_markDev(k, false)
		}
	}
}
func _markDev(name string, dev bool) {
	if _, ok := dependList[name]; !ok {
		return
	}
	dependList[name].IsDev = dev
	for k := range dependList[name].Packages[sel[name]].Package.Require {
		_markDev(k, dev)
	}
}
func installWorker(fileCh <-chan *repositories.JsonPackage, result chan<- int) {
	cacheObj := cache.NewCacheBase()
	for {
		p, ok := <-fileCh
		if !ok {
			return
		}
		t1 := time.Now()
		err := cacheObj.Install(p.Name, p.Dist.Url, p.Dist.Type)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("  - install %s version %s time: %s\r\n", p.Name, p.Version, time.Now().Sub(t1))
		result <- 1
	}
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
		if count == 0 {
			continue
		}
		depSort = append(depSort, &depS{
			name:  name,
			count: count,
		})
	}
	sort.Sort(depSort)
	return depSort
}
