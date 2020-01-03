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
	Shasum    string `json:"shasum"`
}
type JsonSource struct {
	Type      string `json:"type"`
	Url       string `json:"url"`
	Reference string `json:"reference"`
}
type JsonAutoLoad struct {
	Psr4                interface{} `json:"psr-4,omitempty"`
	Psr0                interface{} `json:"psr-0,omitempty"`
	Files               interface{} `json:"files,omitempty"`
	ExcludeFromClassMap interface{} `json:"exclude-from-classmap,omitempty"`
	ClassMap            interface{} `json:"classmap,omitempty"`
}
type JsonPackage struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Source          *JsonSource       `json:"source"`
	Dist            *JsonDist         `json:"dist"`
	Require         map[string]string `json:"require,omitempty"`
	Conflict        map[string]string `json:"conflict,omitempty"`
	Provide         interface{}       `json:"provide,omitempty"`
	Replace         map[string]string `json:"replace,omitempty"`
	RequireDev      map[string]string `json:"require-dev,omitempty"`
	Suggest         interface{}       `json:"suggest,omitempty"`
	Bin             interface{}       `json:"bin,omitempty"`
	Type            string            `json:"type"`
	Extra           interface{}       `json:"extra,omitempty"`
	AutoLoad        *JsonAutoLoad     `json:"autoload,omitempty"`
	NotificationUrl string            `json:"notification-url"`
	License         interface{}       `json:"license"`
	Authors         []struct {
		Name     string `json:"name,omitempty"`
		Email    string `json:"email,omitempty"`
		Homepage string `json:"homepage,omitempty"`
		Role     string `json:"role,omitempty"`
	} `json:"authors,omitempty"`
	Description  string           `json:"description,omitempty"`
	Repositories JsonRepositories `json:"repositories,omitempty"`
	Homepage     string           `json:"homepage,omitempty"`
	Keywords     []string         `json:"keywords,omitempty"`
	Time         string           `json:"time,omitempty"`
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
	Replace *JsonPackage // Package.Package will replaced by Package.Replace
}

type Project struct {
	Constraints map[string]bool
	Conflicts   map[string]bool
	Packages    Packages
	IsDev       bool
}

func (p *Project) GetPackages(index int) *Package {
	return p.Packages[index]
}
func (p *Project) GetRequire(index int) map[string]string {
	return p.Packages[index].Package.Require
}
func (p *Project) GetConflicts(index int) map[string]string {
	return p.Packages[index].Package.Conflict
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
	s1 := strings.Split(ver, "-")
	s := strings.Split(ver, ".")
	if len(s) > 3 {
		ver = strings.Join(s[:3], ".")
		if len(s1) > 1 {
			ver = ver + "-" + s1[1]
		}
	}
	newVersion, err := semver.NewVersion(ver)
	if err != nil || newVersion == nil {
		return nil
	}
	if p == nil {
		p = NewJsonPackage()
	}
	return &Package{newVersion, p, nil}
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
