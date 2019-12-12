package util

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func DownloadAndSave(url, file, hash string) (body []byte, err error) {
	data, err := ioutil.ReadFile(file)
	if len(data) > 0 {
		h := sha256.Sum256(data)
		if hash == hex.EncodeToString(h[:]) {
			return data, nil
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
func Close(i io.Closer) {
	err := i.Close()
	if err != nil {
	}
}
