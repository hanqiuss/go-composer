package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	"sort"
)

type Constraint struct {
	semver.Constraints
	From string
}

type VersionPackage struct {
	v *semver.Version
	p *Package
}
type VersionPackageList []*VersionPackage

func (c VersionPackageList) Len() int           { return len(c) }
func (c VersionPackageList) Less(i, j int) bool { return c[i].v.LessThan(c[j].v) }
func (c VersionPackageList) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

type VersionPack struct {
	Constraints []*Constraint
	Package     []*VersionPackage
}

func FilterVersion(p Packages, vConStr string) *VersionPack {
	constraints, err := semver.NewConstraint(vConStr)
	if err != nil {
		fmt.Println("version error", vConStr)
		return nil
	}
	constraint := Constraint{*constraints, "root"}

	var list VersionPackageList
	for k, v := range p {
		version, err := semver.NewVersion(k)
		if err != nil || version == nil {
			delete(p, k)
			continue
		}

		ok := constraint.Check(version)
		if !ok {
			delete(p, k)
			continue
		}
		list = append(list, &VersionPackage{version, v})
	}
	if list == nil {
		return nil
	}
	var cts []*Constraint
	cts = append(cts, &constraint)
	sort.Sort(sort.Reverse(list))
	return &VersionPack{cts, list}
}
