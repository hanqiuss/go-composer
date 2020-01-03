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
		Conflicts:   map[string]bool{},
		Packages:    repositories.Packages{&repositories.Package{Version: rootVersion, Package: p}},
	}
	dependList = dep
	fmt.Println("start replace at", time.Now())
	checkReplace()
	fmt.Println("start solve at", time.Now())
	phpPkg := &repositories.Package{Version: util.Conf.PhpVer, Package: repositories.NewJsonPackage()}
	phpPkg.Package.Name = "php"
	dependList["php"] = &repositories.Project{
		Constraints: make(map[string]bool),
		Conflicts:   map[string]bool{},
		Packages:    repositories.Packages{phpPkg},
	}
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
		return
	}
	if i == s[n] {
		return
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
	start := 0
	for {
	begin:
		cts := dependList[name].Constraints
		for ccc := range dependList[name].Conflicts {
			cts[ccc] = false
		}
		max := -1
		max2 := 0
		nameList := make([]string, 0)
		ctsList := make(map[*semver.Constraints]bool)
		// get the top index of  the match version
		for depByName, require := range cts {
			if sel[depByName] >= len(dependList[depByName].Packages) {
				return false
			}
			str := ""
			ok := false
			if require {
				str, ok = dependList[depByName].Packages[sel[depByName]].Package.Require[name]
			} else {
				str, ok = dependList[depByName].Packages[sel[depByName]].Package.Conflict[name]

			}
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
			ctsList[curCts] = require
			//ctsList = append(ctsList, curCts)
			nameList = append(nameList, depByName)
			index := getCheckVersionIndex(curCts, ps, start, require)
			if index == len(dependList[name].Packages) {
				continue
			}
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
		for max < n && max >= 0 {
			p := dependList[name].Packages[max]
			check := true
			for ctStr, require := range ctsList {
				if ctStr.Check(p.Version) != require {
					check = false
					break
				}
			}
			if check {
				//is some require make it down but now it can up ? if can't upgrade, start from sel[name]
				if max < sel[name] && !checkPackageRequireOk(dependList[name].Packages[max].Package) {
					start = sel[name]
					goto begin
				}
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
	var failed []string
	for _, depByName := range l {
		endDepBy := len(dependList[depByName].Packages)
		startDepBy := 0
		for startDepBy = sel[depByName]; startDepBy < endDepBy; startDepBy++ {
			require := true
			str, ok := dependList[depByName].GetRequire(startDepBy)[name]
			if !ok {
				require = false
				str = dependList[depByName].GetConflicts(startDepBy)[name]
			}
			str = util.ReWriteConstraint(str)
			if str == "" || str == "*" {
				sel[depByName] = startDepBy
				break
			}
			curCts, err := semver.NewConstraint(str)
			if err != nil {
				continue
			}
			if i := getCheckVersionIndex(curCts, p, start, require); i < end {
				sel[depByName] = startDepBy
				break
			}
		}
		if startDepBy >= endDepBy {
			failed = append(failed, depByName)
			continue
		}
	}
	updateInstallList("root")
	for _, s := range failed {
		if _, ok := dependList[name].Constraints[s]; ok {
			//some require item error
			fmt.Println("_solveMultipleDown error no match ", s, name)
			return false
		}
	}
	return true
}

/*
cts 	version constraint
p       current package list
start   start index
require is require or conflict
*/
func getCheckVersionIndex(cts *semver.Constraints, p repositories.Packages, start int, require bool) int {
	i := len(p)
	for j := start; j < i; j++ {
		if p[j].Version == nil {
			continue
		}
		if cts.CheckPre(p[j].Version) && cts.Check(p[j].Version) == require {
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
		for conflict := range dependList[name].Conflicts {
			str, ok := dependList[conflict].Packages[sel[conflict]].Package.Conflict[name]
			if !ok {
				continue
			}
			str = util.ReWriteConstraint(str)
			ct, err := semver.NewConstraint(str)
			if err != nil {
				fmt.Println("check error : error constraints", str)
				continue
			}
			if ct.Check(ver) {
				return false
			}
		}
	}
	return true
}

func checkPackageRequireOk(p *repositories.JsonPackage) bool {
	for name, vStr := range p.Require {
		if vStr == "" || vStr == "*" {
			continue
		}
		vStr = util.ReWriteConstraint(vStr)
		ct, err := semver.NewConstraint(vStr)
		if err != nil {
			continue
		}
		if !ct.Check(dependList[name].Packages[sel[name]].Version) {
			return false
		}
	}
	return true
}
func updateInstallList(root string) (list map[string]*repositories.JsonPackage) {
	if root == "root" {
		installList = make(map[string]*repositories.JsonPackage)
		for _, p := range dependList {
			p.Constraints = make(map[string]bool)
			p.Conflicts = make(map[string]bool)
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
	for name := range project.GetConflicts(sel[root]) {
		_, ok := dependList[name]
		if !ok {
			continue
		}
		dependList[name].Conflicts[root] = true
	}
	return installList
}
func install() {
	deleteReplace()
	markDev()
	count := 0
	lock := repositories.JsonLock{}
	delete(installList, "root")
	delete(installList, "php")
	numCpu := runtime.NumCPU() * 5
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
	for _, p := range lock.Packages {
		p.NotificationUrl = "https://packagist.org/downloads/"
		sort.Strings(p.Keywords)
	}
	for _, p := range lock.PackagesDev {
		p.NotificationUrl = "https://packagist.org/downloads/"
		sort.Strings(p.Keywords)
	}
	err := util.JsonDataToFile("composer.lock", lock)
	if err != nil {
		fmt.Println("write lock file error ", err)
	}
	if util.Conf.LockOnly {
		return
	}
	lock.Packages = append(lock.Packages, lock.PackagesDev...)
	lock.Sort()
	err = util.JsonDataToFile("vendor/composer/installed.json", lock.Packages)
	if err != nil {
		fmt.Println("write installed.json file error ", err)
	}
	installList["root"] = dependList["root"].Packages[0].Package
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
		if !util.Conf.LockOnly {
			t1 := time.Now()
			err := cacheObj.Install(p.Name, p.Dist.Url, p.Dist.Type)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf("  - install %s version %s time: %s\r\n", p.Name, p.Version, time.Now().Sub(t1))
		}
		result <- 1
	}
}
func checkReplace() {
	for name, project := range dependList {
		for _, pkg := range project.Packages {
			if pkg.Package.Name != name {
				continue
			}
			for replace, c := range pkg.Package.Replace {
				if replace == name {
					continue
				}
				isVer := false
				if c == "self.version" {
					c = pkg.Package.Version
					isVer = true
				}
				constraint, err := semver.NewConstraint(c)
				if err != nil {
					continue
				}

				if _, ok := dependList[replace]; ok {
					for _, pkg2 := range dependList[replace].Packages {
						if constraint.Check(pkg2.Version) {
							if pkg2.Replace != nil {
								goto labelEnd
							}
							pkg2.Replace = pkg.Package
							if isVer {
								goto labelEnd
							}
						}
						if isVer && pkg.Version.GreaterThan(pkg2.Version) {
							break
						}
					}
					/*					if isVer {
										dependList[replace].Packages = append(dependList[replace].Packages, &repositories.Package{
											Version: pkg.Version,
											Package: pkg.Package,
											Replace: pkg.Package,
										})
									}*/
				}
			labelEnd:
			}
		}
	}
	for _, project := range dependList {
		sort.Sort(sort.Reverse(project.Packages))
	}
}
func deleteReplace() {
	for name := range installList {
		if r := dependList[name].Packages[sel[name]].Replace; r != nil {
			if _, ok := installList[r.Name]; ok {
				delete(installList, name)
			}
		}
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
