package util

import (
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

var Conf Config

func init() {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("can't get work dir")
	}
	Conf.Cwd = wd
	Conf.VendorDir = filepath.Join(Conf.Cwd, "vendor")
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use cwd path ")
		getCwd, _ := os.Getwd()
		cacheDir = filepath.Join(getCwd, "cache$$")
	}
	Conf.CacheDir = funcName(cacheDir)
	Conf.Dev = true
}

func funcName(cacheDir string) string {
	return cacheDir
}
