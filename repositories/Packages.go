package repositories

type JsonPackage struct {
	Name        string
	Description string
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string
	Version     string
	Dist        struct {
		Type string
		Url  string
	}
	Source struct {
		Type string
		Url  string
	}
}
type JsonVersionPackages map[string]*JsonPackage
type JsonPackages struct {
	Packages map[string]*JsonVersionPackages // [name:Packages]
}
