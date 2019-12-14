package parse

import (
	"github.com/z7zmey/php-parser/node"
	"github.com/z7zmey/php-parser/node/name"
	"github.com/z7zmey/php-parser/node/stmt"
	"github.com/z7zmey/php-parser/walker"
)

type ClassMap struct {
	Map []string // []className
	Ns  string
}

func (c *ClassMap) EnterNode(w walker.Walkable) bool {
	switch w.(type) {
	case *stmt.Namespace:
		n := w.(*stmt.Namespace)
		if n.NamespaceName == nil {
			return false
		}
		nameNode := n.NamespaceName.(*name.Name).GetParts()
		str := ""
		for _, v := range nameNode {
			str = str + v.(*name.NamePart).Value + "\\\\"
		}
		c.Ns = str
	case *stmt.Interface:
		n := w.(*stmt.Interface)
		n.Stmts = nil
		iName := n.InterfaceName.(*node.Identifier).Value
		c.Map = append(c.Map, c.Ns+iName)
	case *stmt.Class:
		n := w.(*stmt.Class)
		n.Stmts = nil
		cName := n.ClassName.(*node.Identifier).Value
		c.Map = append(c.Map, c.Ns+cName)
	case *stmt.Trait:
		n := w.(*stmt.Trait)
		n.Stmts = nil
		tName := n.TraitName.(*node.Identifier).Value
		c.Map = append(c.Map, c.Ns+tName)
	case *node.Root:

	default:
		return true
	}
	return true
}
func (c ClassMap) LeaveNode(walker.Walkable)              {}
func (c ClassMap) EnterChildNode(string, walker.Walkable) {}
func (c ClassMap) LeaveChildNode(string, walker.Walkable) {}
func (c ClassMap) EnterChildList(string, walker.Walkable) {}
func (c ClassMap) LeaveChildList(string, walker.Walkable) {}
