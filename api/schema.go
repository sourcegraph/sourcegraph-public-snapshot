// GENERATED from sourcegraph.schema - DO NOT EDIT

package api

var Schema = `schema {
	query: Query
}

type Query {
	root: Root
}

type Root {
	repository(id: ID!): Repository
	repositoryByURI(uri: String!): Repository
}

type Repository {
  id: ID!
	uri: String!
	commit(id: String!): Commit
	latest: Commit!
}

type Commit {
	id: String!
	tree(path: String = ""): Tree
}

type Tree {
	directories: [Directory]!
	files: [File]!
}

type Directory {
	name: String!
	tree: Tree!
}

type File {
	name: String!
	content: Blob!
}

type Blob {
	bytes: String!
}
`
