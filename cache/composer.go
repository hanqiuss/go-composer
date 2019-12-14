package cache

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go-composer/util"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Base struct {
	dir         string
	repoDir     string
	filesDir    string
	manifestPre string
}

var ComposerCache Base

func NewCacheBase() *Base {
	if ComposerCache.dir != "" {
		return &ComposerCache
	}
	cacheDir, _ := os.UserCacheDir()
	if cacheDir == "" {
		fmt.Println("can't get home path, will use cwd path ")
		getCwd, _ := os.Getwd()
		cacheDir = filepath.Join(getCwd, "cache$$")
	}
	cacheDir = filepath.Join(cacheDir, "Composer")
	repoDir := filepath.Join(cacheDir, "repo")
	cacheFileDir := filepath.Join(cacheDir, "files")
	_, err := os.Stat(repoDir)
	if err != nil {
		err = os.MkdirAll(repoDir, os.ModePerm)
		if err != nil {
			fmt.Println("make cache dir failed")
			os.Exit(-1)
		}
	}
	_, err = os.Stat(cacheFileDir)
	if err != nil {
		err = os.MkdirAll(repoDir, os.ModePerm)
		if err != nil {
			fmt.Println("make cache dir failed")
			os.Exit(-1)
		}
	}
	ComposerCache = Base{
		dir:         cacheDir,
		repoDir:     repoDir,
		filesDir:    cacheFileDir,
		manifestPre: "provider-",
	}
	return &ComposerCache
}

func (c *Base) GetManifest(name, urlStr string) (r []byte) {
	urlObj, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("cache : urlStr error ", err)
		return
	}
	name = c.manifestPre + strings.ReplaceAll(name, "/", "$") + ".json"

	file := filepath.Join(c.GetRepoDir(urlObj.Host), name)
	r, err = util.DownloadAndSave(urlStr, file, true)
	if err != nil {
		fmt.Println("cache : get manifest error : ", err)
		return
	}
	return r
}
func (c *Base) GetRepoDir(host string) string {
	return filepath.Join(c.repoDir, "https---"+host)
}

func (c *Base) getFilePath(name, url, typ string) string {
	dir := filepath.Join(c.filesDir, name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {

	}
	h := sha1.Sum([]byte(url))
	return filepath.Join(dir, hex.EncodeToString(h[:])) + "." + typ
}
func (c *Base) CreateManifestDir(host string) {
	err := os.MkdirAll(filepath.Join(c.repoDir, "https---"+host), os.ModePerm)
	if err != nil {
	}
}
func (c *Base) Install(name, url, typ string) error {
	if name == "" {
		return fmt.Errorf("install name empty, url : %s type : %s", url, typ)
	}
	file := c.getFilePath(name, url, typ)
	_, err := util.DownloadExist(url, file, false)
	p, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd error %s", err)
	}
	p = filepath.Join(p, "/vendor/"+name)
	switch typ {
	case "zip":
		return Unzip(p, file)
	case util.NpmPkgType:
		return UnTgz(p, file)
	}
	return Unzip(p, file)
}
func Unzip(dir, zipFile string) error {
	err := clearDir(dir)
	if err != nil {
		return err
	}
	f, err := os.Open(zipFile)
	if err != nil {
		return err
	}
	defer util.Close(f)
	info, err := f.Stat()
	if err != nil {
		return err
	}

	z, err := zip.NewReader(f, info.Size())
	if err != nil {
		return fmt.Errorf("unzip %v: %s", zipFile, err)
	}

	// Unzip, enforcing sizes checked earlier.
	for _, zf := range z.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		name := strings.SplitN(zf.Name, "/", 2)
		dst := filepath.Join(dir, name[1])
		if err := os.MkdirAll(filepath.Dir(dst), 0777); err != nil {
			fmt.Println("cache install error : ", filepath.Dir(dst))
			return err
		}
		w, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
		if err != nil {
			return fmt.Errorf("unzip %v: %v", zipFile, err)
		}
		r, err := zf.Open()
		if err != nil {
			util.Close(w)
			return fmt.Errorf("unzip %v: %v", zipFile, err)
		}
		lr := &io.LimitedReader{R: r, N: int64(zf.UncompressedSize64) + 1}
		_, err = io.Copy(w, lr)
		util.Close(r)
		if err != nil {
			util.Close(w)
			return fmt.Errorf("unzip %v: %v", zipFile, err)
		}
		if err := w.Close(); err != nil {
			return fmt.Errorf("unzip %v: %v", zipFile, err)
		}
		if lr.N <= 0 {
			return fmt.Errorf("unzip %v: content too large", zipFile)
		}
	}

	return nil
}
func UnTgz(dir, tgzFile string) error {
	err := clearDir(dir)
	if err != nil {
		return err
	}
	f, err := os.Open(tgzFile)
	if err != nil {
		return err
	}
	defer util.Close(f)
	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer util.Close(gz)
	tr := tar.NewReader(gz)
	for h, err := tr.Next(); err == nil; h, err = tr.Next() {
		fName := strings.Replace(h.Name, "package/", "", 1)
		fName = filepath.Join(dir, fName)
		if err := os.MkdirAll(filepath.Dir(fName), 0777); err != nil {
			fmt.Println("cache install error : ", filepath.Dir(fName))
			return err
		}
		w, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0777)
		if err != nil {
			return fmt.Errorf("ungzip : create file error %v: %v", tgzFile, err)
		}
		_, err = io.Copy(w, tr)
		util.Close(w)
		if err != nil {
			return fmt.Errorf("ungzip : write data error %s", fName)
		}
	}
	return nil
}
func clearDir(dir string) error {
	files, _ := ioutil.ReadDir(dir)
	if len(files) > 0 {
		err := os.RemoveAll(dir)
		if err != nil {
			fmt.Printf("clear path %s error %s\r\n", dir, err)
		}
	}
	if err := os.MkdirAll(dir, 0777); err != nil {
		return err
	}
	return nil
}
