package main

import (
	"encoding/json"
	"fmt"
	"go-composer/cache"
	"go-composer/repositories"
	"go-composer/semver"
	"go-composer/solver"
	"go-composer/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func main() {
	time1 := time.Now()
	/*	parse.Parse()
		return*/
	fmt.Println("start at ", time.Now())
	if util.Conf.ProjectName != "" {
		createProject()
	}

	file := filepath.Join(util.Conf.Cwd, "composer.json")
	p := ReadPackage(file)
	if p == nil {
		fmt.Println("read composer.json error")
		os.Exit(-1)
	}
	solver.Solver(p)
	fmt.Println("end , use time", time.Now().Sub(time1))
}

func ReadPackage(file string) (p *repositories.JsonPackage) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("read composer.json failed")
		return nil
	}
	p = &repositories.JsonPackage{}
	err = json.Unmarshal(data, p)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return p
}

func createProject() {
	cwd := util.Conf.Cwd
	finfo, err := os.Stat(cwd)
	if os.IsNotExist(err) {
		_ = os.MkdirAll(cwd, os.ModePerm)
	} else {
		if !finfo.IsDir() {
			fmt.Println(cwd, " is not dir")
			os.Exit(-1)
		}
		fs, _ := ioutil.ReadDir(cwd)
		if len(fs) > 0 {
			fmt.Println(cwd, " is not empty")
			os.Exit(-1)
		}
	}
	pkgName := util.Conf.ProjectName
	repos := repositories.CreateManager(nil)
	repo := repos["composer"]
	p := repo.GetPackages(pkgName)
	if p == nil {
		fmt.Println("con't find package : ", pkgName)
		os.Exit(-1)
	}
	c := cache.NewCacheBase()
	for _, pkg := range p.Packages {
		if pkg.Package.Require != nil {
			if cts, ok := pkg.Package.Require["php"]; ok {
				cts = util.ReWriteConstraint(cts)
				cts, err := semver.NewConstraint(cts)
				if err != nil { //php版本约束错误
					continue
				}
				if !cts.Check(util.Conf.PhpVer) { //php版本不匹配
					continue
				}
			}
		}
		err = c.InstallPath(pkgName, pkg.Package.Dist.Url, pkg.Package.Dist.Type, cwd)
		if err == nil {
			break
		}
	}
}
