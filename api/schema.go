// GENERATED from sourcegraph.schema - DO NOT EDIT

package api

var Schema = `schema {
	query: Query
}

interface Node {
	id: ID!
}

type Query {
	root: Root
	node(id: ID!): Node
}

type Root {
	repository(uri: String!): Repository
}

type Repository implements Node {
  id: ID!
	uri: String!
	commit(rev: String!): Commit
	latest: Commit!
	branches: [String!]!
	tags: [String!]!
}

type Commit implements Node {
	id: ID!
	sha1: String!
	tree(path: String = "", recursive: Boolean = false): Tree
	languages: [String!]!
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
