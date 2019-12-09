package main

func solver(m MetaData, root string) {
	if m.Require == nil {
		m.Require = make(map[string]string)
	}
	for name, v := range m.RequireDev {
		m.Require[name] = v
	}
	getPackages(m.Require)
}
