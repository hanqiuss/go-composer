package util

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Cwd       string
	VendorDir string
	CacheDir  string
	Dev       bool
}

const NpmPkgType = "tar"

var Conf Config

func init() {
	var cwd string
	flag.StringVar(&cwd, "d", "", "work dir")
	var pro = flag.Bool("pro", false, "install pro packages")

	flag.Parse()
	if cwd != "" {
		Conf.Cwd = cwd
	} else {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Println("can't get work dir")
		}
		Conf.Cwd = wd
	}
	Conf.VendorDir = filepath.Join(Conf.Cwd, "vendor")
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use work dir path ")
		cacheDir = filepath.Join(Conf.Cwd, "cache$$")
	}
	Conf.CacheDir = cacheDir
	Conf.Dev = !*pro
}
