package solver

import "go-composer/repositories"

func Solver(p *repositories.JsonPackage) {
	if p.Require == nil {
		p.Require = make(map[string]string)
	}
	for name, v := range p.RequireDev {
		p.Require[name] = v
	}
	repositories.GetDep(p)
}
