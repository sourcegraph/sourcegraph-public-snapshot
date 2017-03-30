package api

var Schema = `schema {
	query: Query
	mutation: Mutation
}

interface Node {
	id: ID!
}

type Query {
	root: Root!
	node(id: ID!): Node
}

type Root {
	repository(uri: String!): Repository
	repositories(query: String = ""): [Repository!]!
	remoteRepositories: [RemoteRepository!]!
	remoteStarredRepositories: [RemoteRepository!]!
	symbols(id: String!, mode: String!): [Symbol!]!
	currentUser: User
}

type RefFields {
	refLocation: RefLocation
	uri: URI
}

type URI {
	host: String!
	fragment: String!
	path: String!
	query: String!
	scheme: String!
}

type RefLocation {
	startLineNumber: Int!
	startColumn: Int!
	endLineNumber: Int!
	endColumn: Int!
}

type Repository implements Node {
	id: ID!
	uri: String!
	description: String!
	language: String!
	fork: Boolean!
	private: Boolean!
	createdAt: String!
	pushedAt: String!
	commit(rev: String!): CommitState!
	revState(rev: String!): RevState!
	latest: CommitState!
	lastIndexedRevOrLatest: CommitState!
	defaultBranch: String!
	branches: [String!]!
	tags: [String!]!
	expirationDate: Int
}

type Symbol {
	repository: Repository!
	path: String!
	line: Int!
	character: Int!
}

type RemoteRepository {
	uri: String!
	description: String!
	language: String!
	fork: Boolean!
	private: Boolean!
	createdAt: String!
	pushedAt: String!
}

type CommitState {
	commit: Commit
	cloneInProgress: Boolean!
}

type RevState {
	zapRev: ZapRev
	commit: Commit
	cloneInProgress: Boolean!
}

input SearchQuery {
	pattern: String!
	isRegExp: Boolean!
	isWordMatch: Boolean!
	isCaseSensitive: Boolean!
	maxResults: Int!
}

type Commit implements Node {
	id: ID!
	sha1: String!
	tree(path: String = "", recursive: Boolean = false): Tree
	textSearch(query: SearchQuery): [FileMatch!]!
	file(path: String!): File
	languages: [String!]!
}

type ZapRev {
	ref: String!
	base: String!
	branch: String!
}

type CommitInfo {
	rev: String!
	author: Signature
	committer: Signature
	message: String!
}

type Signature {
	person: Person
	date: String!
}

type Person {
	name:  String!
	email: String!
	gravatarHash: String!
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
	content: String!
	blame(startLine: Int!, endLine: Int!): [Hunk!]!
	commits: [CommitInfo!]!
	dependencyReferences(Language: String!, Line: Int!, Character: Int!): DependencyReferences!
}

type FileMatch {
	path: String!
	lineMatches: [LineMatch!]!
}

type LineMatch {
	preview: String!
	lineNumber: Int!
	offsetAndLengths: [[Int!]!]!
}

type DependencyReferences {
	data: String!
}

type Hunk {
	startLine: Int!
	endLine: Int!
	startByte: Int!
	endByte: Int!
	rev: String!
	author: Signature
	message: String!
}

type Organization {
	name: String!
	avatarURL: String!
	description: String!
	collaborators: Int!
}

type Plan {
	name: String!
	cost: Int!
	seats: Int
	renewalDate: Int
	organization: Organization
}

type User {
	githubOrgs: [Organization!]!
	paymentPlan: Plan!
}

type Mutation {
	cancelSubscription(): Boolean!
	updatePaymentSource(tokenID: String!): Boolean!
	subscribeOrg(tokenID: String!, GitHubOrg: String!, seats: Int!): Boolean!
	startOrgTrial(GitHubOrg: String!): Boolean!
}
`
