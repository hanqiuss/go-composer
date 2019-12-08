package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type MetaData struct {
	Name        string
	Description string
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string
}

func ReadComposer(file string) (m MetaData, err error) {
	str, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println("read composer.json failed")
		return m, err
	}

	err = json.Unmarshal(str, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}
