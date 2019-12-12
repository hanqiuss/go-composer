package repositories

import (
	"github.com/Masterminds/semver"
)

type JsonRepos struct {
	Packages         interface{}
	ProvidersUrl     string                   `json:"providers-url"`
	ProviderIncludes map[string]*JsonProvider `json:"provider-includes"`
}
type JsonProvider struct {
	Hash string `json:"sha256"`
}
type JsonProvPack struct {
	Providers map[string]*JsonProvider
}
type JsonPackage struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string `json:"require"`
	Version     string            `json:"version"`
	Dist        struct {
		Type      string `json:"type"`
		Url       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"dist"`
	Source struct {
		Type      string `json:"type"`
		Url       string `json:"url"`
		Reference string `json:"reference"`
	} `json:"source"`
	Type     string                 `json:"type"`
	AutoLoad map[string]interface{} `json:"autoload"`
	License  interface{}            `json:"license"`
}
type JsonVersionPackages map[string]*JsonPackage
type JsonPackages struct {
	Packages map[string]*JsonVersionPackages // [name:Packages]
}
type JsonLock struct {
	Hash        string        `json:"hash"`
	Packages    []JsonPackage `json:"packages"`
	PackagesDev []JsonPackage `json:"packages-dev"`
}
type Package struct {
	Version *semver.Version
	Package *JsonPackage
}
type Packages []*Package
type Project struct {
	Constraints map[string]bool
	Packages    Packages
	Repository  *Composer
}

func (p Packages) Len() int           { return len(p) }
func (p Packages) Less(i, j int) bool { return p[i].Version.LessThan(p[j].Version) }
func (p Packages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
