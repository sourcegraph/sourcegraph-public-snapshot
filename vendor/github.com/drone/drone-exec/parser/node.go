package parser

import "github.com/drone/drone-exec/yaml"

// NodeType identifies the type of a parse tree node.
type NodeType string

// Type returns itself and provides an easy default
// implementation for embedding in a Node. Embedded
// in all non-trivial Nodes.
func (t NodeType) Type() NodeType {
	return t
}

const (
	NodeList    NodeType = "list"
	NodeFilter           = "filter"
	NodeBuild            = "build"
	NodeCache            = "cache"
	NodeClone            = "clone"
	NodeDeploy           = "deploy"
	NodeCompose          = "compose"
	NodeNotify           = "notify"
	NodePublish          = "publish"
)

// Nodes.

type Node interface {
	Type() NodeType
}

// ListNode holds a sequence of nodes.
type ListNode struct {
	NodeType
	Nodes []Node // nodes executed in lexical order.

	// Keys are the map keys (in same order as Nodes), if this list is
	// an ordered map. The values are the Nodes.
	Keys []string
}

// Append appends a node to the list.
func (l *ListNode) append(key string, n Node) {
	l.Nodes = append(l.Nodes, n)
	l.Keys = append(l.Keys, key)
}

func newListNode() *ListNode {
	return &ListNode{NodeType: NodeList}
}

// DockerNode represents a Docker container that
// should be launched as part of the build process.
type DockerNode struct {
	NodeType

	Image       string
	Pull        bool
	Privileged  bool
	Environment []string
	Entrypoint  []string
	Command     []string
	Commands    []string
	Volumes     []string
	ExtraHosts  []string
	Net         string
	AuthConfig  yaml.AuthConfig
	Memory      int64
	CPUSetCPUs  string
	Vargs       map[string]interface{}

	AllowFailure bool
}

func newDockerNode(typ NodeType, c yaml.Container) *DockerNode {
	return &DockerNode{
		NodeType:    typ,
		Image:       c.Image,
		Pull:        c.Pull,
		Privileged:  c.Privileged,
		Environment: c.Environment,
		Entrypoint:  c.Entrypoint,
		Command:     c.Command,
		Volumes:     c.Volumes,
		ExtraHosts:  c.ExtraHosts,
		Net:         c.Net,
		AuthConfig:  c.AuthConfig,
		Memory:      c.Memory,
		CPUSetCPUs:  c.CPUSetCPUs,
	}
}

func newPluginNode(typ NodeType, p yaml.Plugin) *DockerNode {
	node := newDockerNode(typ, p.Container)
	node.Vargs = p.Vargs
	return node
}

func newBuildNode(typ NodeType, b yaml.Build) *DockerNode {
	node := newDockerNode(typ, b.Container)
	node.Commands = b.Commands
	node.AllowFailure = b.AllowFailure
	return node
}

// FilterNode represents a conditional step used to
// filter nodes. If conditions are met the child
// node is executed.
type FilterNode struct {
	NodeType

	Repo    string
	Branch  []string
	Event   []string
	Success string
	Failure string
	Change  string
	Matrix  map[string]string

	Node Node // Node to execution if conditions met
}

func newFilterNode(filter yaml.Filter) *FilterNode {
	return &FilterNode{
		NodeType: NodeFilter,
		Repo:     filter.Repo,
		Branch:   filter.Branch,
		Event:    filter.Event,
		Matrix:   filter.Matrix,
		Success:  filter.Success,
		Failure:  filter.Failure,
		Change:   filter.Change,
	}
}
