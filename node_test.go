package xgraph // import "github.com/orkestr8/xgraph"

type intNode int64

type nodeT struct {
	id         string
	custom     interface{}
	operator   func([]interface{}) (interface{}, error)
	attributes map[string]interface{}
}

func (n *nodeT) OperatorFunc() OperatorFunc {
	return n.operator
}

func (n *nodeT) NodeKey() NodeKey {
	return NodeKey(n.id)
}

func (n *nodeT) String() string {
	return n.id
}

func (n *nodeT) Attributes() map[string]interface{} {
	return n.attributes
}
