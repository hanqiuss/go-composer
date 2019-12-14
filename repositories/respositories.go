package repositories

var manager = make(map[string]Repository)

func CreateManager(jsonPackage *JsonPackage) map[string]Repository {
	manager["composer"] = NewComposer("")
	manager["npm"] = NewNpm("")
	for range jsonPackage.Repositories {

	}
	return manager
}
