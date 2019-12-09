package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {

	var file, err = os.Getwd()
	if err != nil {
		os.Exit(-1)
	}
	file = filepath.Join(file, "composer.json")
	metadata, err := ReadComposer(file)
	if err != nil {
		fmt.Println(err)
	}
	solver(metadata, "root")

}
