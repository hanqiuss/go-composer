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
		nameNode := n.NamespaceName.(*name.Name).GetParts()
		str := ""
		for _, v := range nameNode {
			str = str + v.(*name.NamePart).Value + "\\\\"
		}
		c.Ns = str
	case *stmt.Interface:
		n := w.(*stmt.Interface)
		n.Stmts = nil
		name := n.InterfaceName.(*node.Identifier).Value
		c.Map = append(c.Map, c.Ns+name)
	case *stmt.Class:
		n := w.(*stmt.Class)
		n.Stmts = nil
		name := n.ClassName.(*node.Identifier).Value
		c.Map = append(c.Map, c.Ns+name)
	case *node.Root:

	default:
		return false
	}
	return true
}
func (c ClassMap) LeaveNode(w walker.Walkable)                  {}
func (c ClassMap) EnterChildNode(key string, w walker.Walkable) {}
func (c ClassMap) LeaveChildNode(key string, w walker.Walkable) {}
func (c ClassMap) EnterChildList(key string, w walker.Walkable) {}
func (c ClassMap) LeaveChildList(key string, w walker.Walkable) {}
