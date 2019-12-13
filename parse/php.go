package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/yookoala/realpath"
	"github.com/z7zmey/php-parser/parser"
	"github.com/z7zmey/php-parser/php7"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var wg sync.WaitGroup

type file struct {
	path    string
	content []byte
}
type Dir struct {
	Path    string
	Exclude []string
}

var classMap = make(map[string]string)

func Parse(dirs []Dir) map[string]string {
	numCpu := runtime.NumCPU()
	fileCh := make(chan *file, numCpu)
	resultCh := make(chan parser.Parser, numCpu)
	// run 4 concurrent parserWorkers
	for i := 0; i < numCpu; i++ {
		go parserWorker(fileCh, resultCh)
	}
	// run printer goroutine
	go printerWorker(resultCh)

	processPath(dirs, fileCh)

	// wait the all files done
	wg.Wait()
	close(fileCh)
	close(resultCh)
	return classMap
}

func processPath(pathList []Dir, fileCh chan<- *file) {
	for _, path := range pathList {
		realP, err := realpath.Realpath(path.Path)
		checkErr(err)

		err = filepath.Walk(realP, func(fPath string, f os.FileInfo, err error) error {
			if f.IsDir() && len(path.Exclude) > 0 {
				fmt.Println(fPath, filepath.Join(path.Path, path.Exclude[0]))
			}
			if !f.IsDir() && filepath.Ext(fPath) == ".php" {
				wg.Add(1)
				content, err := ioutil.ReadFile(fPath)
				checkErr(err)
				fileCh <- &file{fPath, content}
			}
			return nil
		})
		checkErr(err)
	}
}

func parserWorker(fileCh <-chan *file, result chan<- parser.Parser) {
	var parserWorker parser.Parser

	for {
		f, ok := <-fileCh
		if !ok {
			return
		}

		src := bytes.NewReader(f.content)

		parserWorker = php7.NewParser(src, f.path)

		parserWorker.Parse()

		result <- parserWorker
	}
}

func printerWorker(result <-chan parser.Parser) {

	w := bufio.NewWriter(os.Stdout)
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("php.go get cwd error", err)
	}
	wd = filepath.Join(wd, "vendor")
	for {
		parserWorker, ok := <-result
		if !ok {
			err := w.Flush()
			if err != nil {
			}
			return
		}

		path := parserWorker.GetPath()
		path = strings.Replace(path, wd, "", 1)
		path = strings.ReplaceAll(path, "\\", "/")
		var cm = &ClassMap{make([]string, 0), ""}
		parserWorker.GetRootNode().Walk(cm)
		if len(cm.Map) > 0 {
			for _, v := range cm.Map {
				classMap[v] = path
			}
		}
		wg.Done()
	}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}