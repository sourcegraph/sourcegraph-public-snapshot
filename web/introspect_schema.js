const { graphql, introspectionQuery, buildSchema } = require('graphql')
const fs = require('fs')

// Reads the GraphQL IDL schema and generates a JSON introspection query result
// that the GraphQL TS language service plugin can use

async function main() {
  const schemaStr = fs.readFileSync(__dirname + '/../cmd/frontend/internal/graphqlbackend/schema.graphql', 'utf8')
  const schema = buildSchema(schemaStr)
  const result = await graphql(schema, introspectionQuery)
  const json = JSON.stringify(result, null, 2)
  fs.writeFileSync(__dirname + '/graphqlschema.json', json)
}

main()
