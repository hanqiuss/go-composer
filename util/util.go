package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

func DownloadAndSave(url, file string, needReturn bool) (body []byte, err error) {
	fInfo, err := os.Stat(file)
	if err == nil {
		resp, err := http.Head(url)
		if err == nil && resp.StatusCode == 200 {
			d, err := http.ParseTime(resp.Header.Get("Date"))
			if err == nil && !d.After(fInfo.ModTime().Add(time.Hour*2)) {
				if needReturn {
					return ioutil.ReadFile(file)
				}
				return body, nil
			}
		}
	}
	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		return body, fmt.Errorf("get url error %s %s", url, err)
	}
	body, err = ioutil.ReadAll(res.Body)
	defer Close(res.Body)
	if err != nil {
		return body, fmt.Errorf("read http body error %s", err)
	}
	err = ioutil.WriteFile(file, body, os.ModePerm)
	if err != nil {
		return body, fmt.Errorf("write file error %s", err)
	}
	return
}
func DownloadExist(url, file string, needReturn bool) (body []byte, err error) {
	_, err = os.Stat(file)
	if err == nil {
		if needReturn {
			return ioutil.ReadFile(file)
		}
		return body, nil
	}
	//fmt.Println("get url", url)
	res, err := http.Get(url)
	if err != nil || res.StatusCode != 200 {
		return body, fmt.Errorf("get url error %s %s", url, err)
	}
	body, err = ioutil.ReadAll(res.Body)
	defer Close(res.Body)
	if err != nil {
		return body, fmt.Errorf("read http body error %s", err)
	}
	err = ioutil.WriteFile(file, body, os.ModePerm)
	if err != nil {
		return body, fmt.Errorf("write file error %s", err)
	}
	return
}
func Close(i io.Closer) {
	err := i.Close()
	if err != nil {
	}
}

func MD5ToString(b []byte) string {
	h := md5.Sum(b)
	return hex.EncodeToString(h[:])
}
func ReWriteConstraint(v string) string {
	v = strings.ReplaceAll(strings.ReplaceAll(v, "||", "|"), "|", "||")
	return strings.ReplaceAll(v, "@", "-")
}
func FilterRequire(name string) bool {
	ok, _ := regexp.MatchString("^php$", name)
	if ok {
		return false
	}
	ok, _ = regexp.MatchString("^(ext|lib)-.*", name)
	if ok {
		return false
	}
	if !strings.Contains(name, "/") {
		return false
	}
	return true
}

func JsonDataToFile(f string, data interface{}) error {
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json encode error %s", err)
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "    ")
	if err != nil {
		return fmt.Errorf("json Indent error %s", err)
	}
	_ = os.MkdirAll(filepath.Dir(f), os.ModePerm)
	err = ioutil.WriteFile(f, buf.Bytes(), os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}
