package repositories

import (
	"fmt"
	"net/url"
	"sync"
)

var manager = make(map[string]Repository)

func CreateManager(jsonPackage *JsonPackage) map[string]Repository {
	manager["composer"] = NewComposer("")

	count := 0
	var wg sync.WaitGroup
	var mut sync.Mutex
	jsonPackage.Repositories = append(jsonPackage.Repositories, JsonRepository{
		Type: "composer",
		Url:  "",
	})
	for _, repo := range jsonPackage.Repositories {
		switch repo.Type {
		case "composer":
			u, err := url.Parse(repo.Url)
			if err != nil {
				fmt.Println(err)
				continue
			}
			name := "composer:" + u.Host
			_, ok := manager[name]
			if ok {
				continue
			}
			wg.Add(1)
			go func(name, url string) {
				c := NewComposer(url)
				mut.Lock()
				manager[name] = c
				mut.Unlock()
				wg.Done()
			}(name, repo.Url)
			count++
		}
	}
	wg.Wait()
	return manager
}
