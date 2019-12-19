package repositories

import (
	"go-composer/semver"
	"go-composer/util"
	"sort"
	"strings"
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
type JsonDist struct {
	Type      string `json:"type"`
	Url       string `json:"url"`
	Reference string `json:"reference"`
}
type JsonSource JsonDist
type JsonPackage struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Source       *JsonSource            `json:"source"`
	Dist         *JsonDist              `json:"dist"`
	Require      map[string]string      `json:"require"`
	RequireDev   map[string]string      `json:"require-dev"`
	Type         string                 `json:"type"`
	Extra        interface{}            `json:"extra"`
	AutoLoad     map[string]interface{} `json:"autoload"`
	License      interface{}            `json:"license"`
	Description  string                 `json:"description"`
	Repositories JsonRepositories       `json:"repositories"`
}

type JsonVersionPackages map[string]*JsonPackage
type JsonPackages struct {
	Packages map[string]*JsonVersionPackages // [name:Packages]
}
type JsonLock struct {
	Hash        string     `json:"hash"`
	Packages    ColJsonPkg `json:"packages"`
	PackagesDev ColJsonPkg `json:"packages-dev"`
}

func (l *JsonLock) Sort() {
	sort.Sort(l.Packages)
	sort.Sort(l.PackagesDev)
}

type JsonRepository struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}
type JsonRepositories interface{}
type Package struct {
	Version *semver.Version
	Package *JsonPackage
}

type Project struct {
	Constraints map[string]bool
	Packages    Packages
	IsDev       bool
}

func (p *Project) GetPackages(index int) *Package {
	return p.Packages[index]
}
func (p *Project) GetRequire(index int) map[string]string {
	return p.Packages[index].Package.Require
}

type JsonNpmPackage struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	RequireDev  map[string]string `json:"require-dev"`
	Require     map[string]string `json:"dependencies"`
	Version     string            `json:"version"`
	GitHead     string            `json:"gitHead"`
	Dist        struct {
		Type   string `json:"type"`
		Url    string `json:"tarball"`
		ShaSum string `json:"shasum"`
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

func NewJsonPackage() *JsonPackage {
	return &JsonPackage{}
}
func NewPackage(ver string, p *JsonPackage) *Package {
	s := strings.Split(ver, ".")
	if len(s) > 3 {
		ver = strings.Join(s[:3], ".")
	}
	newVersion, err := semver.NewVersion(ver)
	if err != nil || newVersion == nil {
		return nil
	}
	if p == nil {
		p = NewJsonPackage()
	}
	return &Package{newVersion, p}
}

type ColJsonPkg []*JsonPackage

func (p ColJsonPkg) Len() int           { return len(p) }
func (p ColJsonPkg) Less(i, j int) bool { return p[i].Name < p[j].Name }
func (p ColJsonPkg) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type Packages []*Package

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
		Dist: &JsonDist{
			Type:      util.NpmPkgType,
			Url:       npm.Dist.Url,
			Reference: npm.GitHead,
		},
		Source: &JsonSource{
			Type:      npm.Source.Type,
			Url:       strings.Replace(npm.Source.Url, "git+http", "http", 1),
			Reference: npm.GitHead,
		},
		Type:         "",
		AutoLoad:     nil,
		License:      npm.License,
		Repositories: nil,
	}
}
