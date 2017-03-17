package api

//go:generate ./gen/gen-go.sh
//go:generate go run ./gen/gen-json.go
//go:generate yarn install
//go:generate ./node_modules/.bin/gql2ts schema.json -o ./graphqlInterfaces.gen.d.ts
//go:generate ./node_modules/.bin/tsfmt -r ./graphqlInterfaces.gen.d.ts
