package repositories

var manager map[string]Repository

func CreateManager(jsonPackage *JsonPackage) map[string]Repository {
	if manager != nil {
		return manager
	}
	manager = make(map[string]Repository)
	manager["composer"] = NewComposer("")
	manager["npm"] = NewNpm("")

	return manager
}
