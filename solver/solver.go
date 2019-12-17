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

	for !checkDep() {
		if !solveDep() {
			fmt.Println("solver error")
			return
		}
	}
	fmt.Println("end solver at", time.Now())
	install()
	fmt.Println("end install at", time.Now())
}

type selected map[string]int

func (s selected) set(n string, i int) {
	if i >= len(dependList[n].Packages) {
		s[n] = len(dependList[n].Packages)
	}
	old := s[n]
	r1 := dependList[n].GetRequire(old)
	r2 := dependList[n].GetRequire(i)
	update := false
	for dep := range r1 { //dep in r1 but not in r2
		if _, ok := r2[dep]; !ok {
			if _, ok := dependList[dep]; ok {
				update = true
			}
		}
	}
	for dep := range r2 { //dep in r2 but not in r1
		if _, ok := r1[dep]; !ok {
			if _, ok := dependList[dep]; ok {
				update = true
			}
		}
	}
	if update {
		updateInstallList("root")
	}
	s[n] = i
}
func (s selected) add(n string) {
	s.set(n, s[n]+1)
}

var sel = selected{}
var dependList map[string]*repositories.Project
var installList map[string]*repositories.JsonPackage

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
	for {
	begin:
		cts := dependList[name].Constraints
		max := 0
		max2 := 0
		nameList := make([]string, 0)
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
				sel.add(depByName)
				fmt.Println("error constraints ", str, depByName)
				goto begin
			}
			ps := dependList[name].Packages
			index := getCheckVersionIndex(curCts, ps, 0)
			if index == len(dependList[name].Packages) {
				fmt.Printf("package %s need %s %s, no match\r\n", depByName, name, str)
				sel.add(depByName)
				goto begin
			}
			ctsList = append(ctsList, curCts)
			nameList = append(nameList, depByName)
			if max < index {
				max = index
			}
		}
		if len(ctsList) == 0 {
			return true
		}
		n := len(dependList[name].Packages)
		// check is some version match all constraints
		max2 = max
		for max < n {
			p := dependList[name].Packages[max]
			check := true
			for _, cts := range ctsList {
				if !cts.Check(p.Version) {
					check = false
					break
				}
			}
			if check {
				sel.set(name, max)
				return true
			}
			max++
		}
		// downgrade the dependBy version which has the max require version(constraints)
		_solveMultipleDown(nameList, name, max2)
	}
}
func _solveMultipleDown(l []string, name string, start int) bool {
	if len(l) == 0 {
		fmt.Println("_solveMultipleDown error list ", name)
		return false
	}
	if len(l) == 1 {
		sel.add(l[0])
		return true
	}
	p := dependList[name].Packages
	end := len(p)
	for _, depByName := range l {
		endDepBy := len(dependList[depByName].Packages)
		startDepBy := 0
		for startDepBy = sel[depByName]; startDepBy < endDepBy; startDepBy++ {
			str := dependList[depByName].Packages[startDepBy].Package.Require[name]
			str = util.ReWriteConstraint(str)
			if str == "" || str == "*" {
				sel.set(depByName, startDepBy)
				break
			}
			curCts, err := semver.NewConstraint(str)
			if err != nil {
				continue
			}
			if i := getCheckVersionIndex(curCts, p, start); i < end {
				sel.set(depByName, startDepBy)
				break
			}
		}
		if startDepBy >= endDepBy {
			fmt.Println("_solveMultipleDown error no match ", depByName, name)
			return false
		}
	}
	return true
}
func getCheckVersionIndex(cts *semver.Constraints, p repositories.Packages, start int) int {
	i := len(p)
	for j := start; j < i; j++ {
		if p[j].Version == nil {
			continue
		}
		if cts.Check(p[j].Version) {
			i = j
		}
	}
	return i
}
func checkDep() bool {
	updateInstallList("root")
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
func updateInstallList(root string) (list map[string]*repositories.JsonPackage) {
	if root == "root" {
		installList = make(map[string]*repositories.JsonPackage)
		for _, p := range dependList {
			p.Constraints = make(map[string]bool)
		}
	}
	project, ok := dependList[root]
	if !ok {
		return
	}
	for name := range project.GetRequire(sel[root]) {
		p, ok := dependList[name]
		if !ok {
			continue
		}
		dependList[name].Constraints[root] = true
		if _, ok := installList[name]; ok {
			continue
		}
		installList[name] = p.Packages[sel[name]].Package
		updateInstallList(name)
	}
	return installList
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

var markMap = make(map[string]bool)

func markDev() {
	if !util.Conf.Dev {
		return
	}
	root := dependList["root"].Packages[0].Package
	markMap = make(map[string]bool)
	for k := range root.RequireDev {
		_markDev(k, true)
	}
	markMap = make(map[string]bool)
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
	_, ok := markMap[name]
	if ok {
		return
	}
	dependList[name].IsDev = dev
	markMap[name] = true
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
