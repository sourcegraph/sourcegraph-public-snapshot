package api

//go:generate ./gen/gen-go.sh
//go:generate go run ./gen/gen-json.go
//go:generate ../ui/node_modules/.bin/gql2ts schema.json -o ../ui/web_modules/sourcegraph/graphqlInterfaces.gen.d.ts
