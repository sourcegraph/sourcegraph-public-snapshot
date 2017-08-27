package graphqlbackend

//go:generate go run ./gen/gen-json.go
//go:generate npm install
//go:generate ./node_modules/.bin/gql2ts ../../../../client/schema.json -o ../../../../client/graphqlInterfaces.gen.d.ts
//go:generate ./node_modules/.bin/tsfmt -r ../../../../client/graphqlInterfaces.gen.d.ts
