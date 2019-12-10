package main

import (
	"encoding/json"
	"fmt"
	"go-composer/repositories"
	"go-composer/solver"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {

	var file, err = os.Getwd()
	if err != nil {
		os.Exit(-1)
	}
	file = filepath.Join(file, "composer.json")
	p := ReadPackage(file)
	if p == nil {
		fmt.Println("read composer.json error")
		os.Exit(-1)
	}
	solver.Solver(p)

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
