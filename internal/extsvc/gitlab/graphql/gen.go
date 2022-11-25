//go:build generate

//go:generate go run github.com/Khan/genqlient
package graphql

import _ "github.com/Khan/genqlient"
