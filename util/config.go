package util

import (
	"flag"
	"fmt"
	"go-composer/semver"
	"os"
	"os/exec"
	"path/filepath"
)

type Config struct {
	Cwd       string
	VendorDir string
	CacheDir  string
	Dev       bool
	PhpVer    *semver.Version
	LockOnly  bool
}

const NpmPkgType = "tar"

var Conf Config

func init() {
	var cwd, phpV string
	flag.StringVar(&cwd, "d", "", "work dir")
	flag.StringVar(&phpV, "php", "", "php version")
	var pro = flag.Bool("pro", false, "install pro packages")
	var lock = flag.Bool("lockonly", false, "only make .lock file")
	flag.Parse()
	getPhpVer(phpV)
	getCwd(cwd)
	getCacheDir()
	Conf.LockOnly = *lock
	Conf.Dev = !*pro
}

func getCwd(cwd string) {
	if cwd != "" {
		Conf.Cwd = cwd
	} else {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("can't get work dir")
		}
		Conf.Cwd = wd
	}
}
func getCacheDir() {
	Conf.VendorDir = filepath.Join(Conf.Cwd, "vendor")
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use work dir path ")
		cacheDir = filepath.Join(Conf.Cwd, "cache$$")
	}
	Conf.CacheDir = cacheDir
}
func getPhpVer(v string) {
	if v != "" {
		ver, err := semver.NewVersion(v)
		if err != nil {
			fmt.Println("php version error", err)
		} else {
			Conf.PhpVer = ver
		}

	}
	if Conf.PhpVer == nil {
		cmd := exec.Command("php", "-r", `echo PHP_VERSION;`)
		r, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
		ver, err := semver.NewVersion(string(r))
		if err != nil {
			fmt.Println("php version error", err, string(r))
		} else {
			Conf.PhpVer = ver
		}
	}
	if Conf.PhpVer == nil {
		fmt.Println("unknown php version, with -php={version} in cmd or add the exec path to env")
		os.Exit(-1)
	} else {
		fmt.Println("start with php version : ", Conf.PhpVer)
	}
}
