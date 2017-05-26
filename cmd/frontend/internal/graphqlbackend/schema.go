package graphqlbackend

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
	organization(login: String!): Organization
	repository(uri: String!): Repository
	repositories(query: String = ""): [Repository!]!
	remoteRepositories: [RemoteRepository!]!
	remoteStarredRepositories: [RemoteRepository!]!
	symbols(id: String!, mode: String!): [Symbol!]!
	currentUser: User
	searchRepos(query: SearchQuery!, repositories: [RepositoryRevision!]!): SearchResults!
	revealCustomerCompany(ip: String!): CompanyProfile
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
	gitCmdRaw(params: [String!]!): String!
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
	commit: Commit
	cloneInProgress: Boolean!
}

input SearchQuery {
	pattern: String!
	isRegExp: Boolean!
	isWordMatch: Boolean!
	isCaseSensitive: Boolean!
	fileMatchLimit: Int!
	includePattern: String
	excludePattern: String
}

input RepositoryRevision {
	repo: String!
	rev: String	
}

type Commit implements Node {
	id: ID!
	sha1: String!
	tree(path: String = "", recursive: Boolean = false): Tree
	textSearch(query: SearchQuery): SearchResults!
	file(path: String!): File
	languages: [String!]!
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
	blameRaw(startLine: Int!, endLine: Int!): String!
}

type SearchResults {
	results: [FileMatch!]!
	limitHit: Boolean!
}

type FileMatch {
	resource: String!
	lineMatches: [LineMatch!]!
	limitHit: Boolean!
}

type LineMatch {
	preview: String!
	lineNumber: Int!
	offsetAndLengths: [[Int!]!]!
	limitHit: Boolean!
}

type DependencyReferences {
	dependencyReferenceData: DependencyReferencesData!
	repoData: RepoDataMap!
}

type RepoDataMap {
	repos: [Repository!]!
	repoIds: [Int!]!
}

type DependencyReferencesData {
	references: [DependencyReference!]!
	location: DepLocation!
}

type DependencyReference {
	dependencyData: String!
	repoId: Int!
	hints: String!
}

type DepLocation {
	location: String!
	symbol: String!
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
	login: String!
	githubId: Int!
	email: String!
	name: String!
	avatarURL: String!
	description: String!
	collaborators: Int!
	members: [OrganizationMember!]!
}

type OrganizationMember {
	login: String!
	githubId: Int!
	email: String!
	avatarURL: String!
	isSourcegraphUser: Boolean!
	canInvite: Boolean!
	invite: Invite
}

type Invite {
	userLogin: String!
	userEmail: String!
	orgGithubId: Int!
	orgLogin: String!
	sentAt: Int!
	uri: String!
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
	inviteOrgMemberToSourcegraph(orgLogin: String!, orgGithubId: Int!, userLogin: String!, userEmail: String = ""): Boolean!
}

type CompanyProfile {
	ip: String!
	domain: String!
	fuzzy: Boolean!
	company: CompanyInfo!
}

type CompanyInfo {
	id: String!
	name: String!
	legalName: String!
	domain: String!
	domainAliases: [String!]!
	url: String!
	site: SiteDetails!
	category: CompanyCategory!
	tags: [String!]!
	description: String!
	foundedYear: String!
	location: String!
	logo: String!
	tech: [String!]!
}

type SiteDetails {
	url: String!
	title: String!
	phoneNumbers: [String!]!
	emailAddresses: [String!]!
}

type CompanyCategory {
	sector: String!
	industryGroup: String!
	industry: String!
	subIndustry: String!
}
`
