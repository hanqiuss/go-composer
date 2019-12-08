package main

import "fmt"

func solver(m MetaData, root bool, dev bool) {
	fmt.Println(m.Require, m.RequireDev)
	if m.Require == nil {
		m.Require = make(map[string]string)
	}
	if root && dev {
		for name, v := range m.RequireDev {
			m.Require[name] = v
		}
	}
	getPackages(m.Require)
}
