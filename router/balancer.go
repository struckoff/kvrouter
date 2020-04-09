package router

type Balancer interface {
	AddNode(n Node) error
	RemoveNode(id string) error
	SetNodes(ns []Node) error
	LocateKey(key string) (Node, error)
	Nodes() ([]Node, error)
	GetNode(id string) (Node, error)
}
