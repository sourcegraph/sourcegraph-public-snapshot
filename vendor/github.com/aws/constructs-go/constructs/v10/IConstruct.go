package constructs

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

// Represents a construct.
type IConstruct interface {
	IDependable
	// The tree node.
	Node() Node
}

// The jsii proxy for IConstruct
type jsiiProxy_IConstruct struct {
	jsiiProxy_IDependable
}

func (j *jsiiProxy_IConstruct) Node() Node {
	var returns Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

