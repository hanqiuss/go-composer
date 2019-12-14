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
	Type         string                 `json:"type"`
	AutoLoad     map[string]interface{} `json:"autoload"`
	License      interface{}            `json:"license"`
	Repositories JsonRepositories       `json:"repositories"`
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
type JsonRepository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}
type JsonRepositories []JsonRepository
type Package struct {
	Version *semver.Version
	Package *JsonPackage
}
type Packages []*Package
type Project struct {
	Constraints map[string]bool
	Packages    Packages
}

type JsonNpmPackage struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string `json:"dependencies"`
	Version     string            `json:"version"`
	Dist        struct {
		Type      string `json:"type"`
		Url       string `json:"tarball"`
		Reference string `json:"reference"`
	} `json:"dist"`
	Source struct {
		Type string `json:"type"`
		Url  string `json:"url"`
	} `json:"repository"`
	License interface{} `json:"license"`
}
type JsonNpmVersionPackage map[string]*JsonNpmPackage
type JsonNpmPackages struct {
	Versions JsonNpmVersionPackage
}

func (p Packages) Len() int           { return len(p) }
func (p Packages) Less(i, j int) bool { return p[i].Version.LessThan(p[j].Version) }
func (p Packages) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Repository interface {
	GetPackages(name string) *Project
	Has(name string) bool
}

func JsonNpmToComposer(npm *JsonNpmPackage) *JsonPackage {
	return &JsonPackage{
		Name:        npm.Name,
		Description: npm.Version,
		RequireDev:  npm.RequireDev,
		Require:     npm.Require,
		Version:     npm.Version,
		Dist: struct {
			Type      string `json:"type"`
			Url       string `json:"url"`
			Reference string `json:"reference"`
		}{
			Type:      "tgz",
			Url:       npm.Dist.Url,
			Reference: "",
		},
		Source: struct {
			Type      string `json:"type"`
			Url       string `json:"url"`
			Reference string `json:"reference"`
		}{
			Type:      npm.Source.Type,
			Url:       npm.Source.Url,
			Reference: "",
		},
		Type:         "",
		AutoLoad:     nil,
		License:      npm.License,
		Repositories: nil,
	}
}
