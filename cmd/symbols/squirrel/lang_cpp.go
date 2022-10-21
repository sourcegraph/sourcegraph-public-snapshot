package squirrel

import (
	"context"
	"fmt"
)

func (squirrel *SquirrelService) getDefCpp(ctx context.Context, node Node) (ret *Node, err error) {
	fmt.Printf("NODE %v\n", node)
	fmt.Printf("NODE    %v\n", node.Type())
	fmt.Printf("PARENT  %v\n", node.Parent().Type())
	fmt.Printf("PARENT2 %v\n", node.Parent().Parent().Type())
	fmt.Printf("NODE %v\n", getStringContents(node))
	return nil, nil
}
