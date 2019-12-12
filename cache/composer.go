package cache

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"go-composer/util"
	"io"
	"io/ioutil"
	"net/http"
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
	repoDir := filepath.Join(cacheDir, "repo", "https---repo.packagist.org")
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
func (c *Base) GetManifest(name, url string) (r []byte) {
	file := c.getManifestPath(name)
	info, err := os.Stat(file)
	if err == nil {
		resp, err := http.Head(url)
		if err != nil {
			fmt.Println("get url failed : ", url)
		} else {
			timeStr := resp.Header.Get("date")
			if timeStr == "" { //read from url
				return
			}
			time, err := http.ParseTime(timeStr)
			if err != nil {
				return nil
			}
			if time.After(info.ModTime()) {
				return
			}
		}
		ret, err := ioutil.ReadFile(file)
		if err != nil {
			return
		}
		return ret
	}
	return
}
func (c *Base) getManifestPath(name string) string {
	return filepath.Join(c.repoDir, c.manifestPre+strings.ReplaceAll(name, "/", "$")) + ".json"
}
func (c *Base) GetRepoDir() string {
	return c.repoDir
}
func (c *Base) CacheManifest(name string, body []byte) bool {
	file := c.getManifestPath(name)
	err := ioutil.WriteFile(file, body, os.ModePerm)
	return err == nil
}
func (c *Base) CacheFiles(name, url, typ string) bool {
	file := c.getFilePath(name, url, typ)
	_, err := os.Stat(file)
	if err == nil {
		return true
	}
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return false
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			return
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	err = ioutil.WriteFile(file, body, os.ModePerm)
	return err == nil
}
func (c *Base) GetFiles(name, url, typ string) []byte {
	file := c.getFilePath(name, url, typ)
	_, err := os.Stat(file)
	if err != nil {
		c.CacheFiles(name, url, typ)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return []byte{}
	}
	return data
}
func (c *Base) getFilePath(name, url, typ string) string {
	dir := filepath.Join(c.filesDir, name)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {

	}
	h := sha1.Sum([]byte(url))
	return filepath.Join(dir, hex.EncodeToString(h[:])) + "." + typ
}

func (c *Base) Install(name, url, typ string) error {
	if name == "" {
		return fmt.Errorf("install name empty, url : %s type : %s", url, typ)
	}
	file := c.getFilePath(name, url, typ)
	p, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get cwd error %s", err)
	}
	p = filepath.Join(p, "/vendor/"+name)
	return Unzip(p, file)
}
func Unzip(dir, zipFile string) error {

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
			fmt.Println("123", filepath.Dir(dst))
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
