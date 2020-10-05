package graphs

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

func graphURL(o graphqlbackend.GraphOwner, g graphqlbackend.GraphResolver) string {
	return o.URL() + "/graphs/" + string(g.Name())
}
