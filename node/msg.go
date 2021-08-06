package node

var nc *NodeContext

func Current() *NodeContext {
	return nc
}

func SetNodeContext(ctx *NodeContext) {
	nc = ctx
}
