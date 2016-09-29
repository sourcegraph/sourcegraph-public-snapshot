// GENERATED CODE - DO NOT EDIT!

export interface Annotation {
	URL?: string;
	StartByte: number;
	EndByte: number;
	Class?: string;
	Def?: boolean;
	WantInner?: number;
	URLs?: string[];
}

export interface AnnotationList {
	Annotations?: Annotation[];
	LineStartBytes?: number[];
}

export interface AnnotationsGetDefAtPosOptions {
	Entry: TreeEntrySpec;
	Line?: number;
	Character?: number;
}

export interface AnnotationsListOptions {
	Entry: TreeEntrySpec;
	Range?: FileRange;
	NoSrclibAnns?: boolean;
}

export interface AsyncRefreshIndexesOp {
	Repo?: number;
	Source?: string;
	Force?: boolean;
}

export interface AuthInfo {
	UID?: string;
	Login?: string;
	Write?: boolean;
	Admin?: boolean;
}

export interface AuthorshipInfo {
	LastCommitDate: any;
	LastCommitID?: string;
}

export interface BasicTreeEntry {
	Name?: string;
	Type?: any;
	CommitID?: string;
	Contents?: number[];
	Entries?: BasicTreeEntry[];
}

export interface BehindAhead {
	Behind?: number;
	Ahead?: number;
}

export interface BetaRegistration {
	Email?: string;
	FirstName?: string;
	LastName?: string;
	Languages?: string[];
	Editors?: string[];
	Message?: string;
}

export interface BetaResponse {
	EmailAddress?: string;
}

export interface BlameOptions {
}

export interface Branch {
	Name?: string;
	Head?: any;
	Commit?: any;
	Counts?: any;
}

export interface BranchList {
	Branches?: any[];
	HasMore?: boolean;
}

export interface BranchesOptions {
	MergedInto?: string;
	IncludeCommit?: boolean;
	BehindAheadBranch?: string;
	ContainsCommit?: string;
}

export interface Change {
}

export interface Client {
}

export interface ClientCapabilities {
}

export interface CodeActionContext {
	diagnostics: any[];
}

export interface CodeActionParams {
	textDocument: any;
	range: any;
	context: any;
}

export interface CodeLens {
	range: any;
	command?: any;
	data?: any;
}

export interface CodeLensOptions {
	resolveProvider?: boolean;
}

export interface CodeLensParams {
	textDocument: any;
}

export interface CombinedStatus {
	Rev?: string;
	CommitID?: string;
	State?: string;
	Statuses?: RepoStatus[];
}

export interface Command {
	title: string;
	command: string;
	arguments: any[];
}

export interface Commit {
	ID?: any;
	Author: any;
	Committer?: any;
	Message?: string;
	Parents?: any[];
}

export interface CommitList {
	Commits?: any[];
	HasMore?: boolean;
}

export interface CommitsOptions {
}

export interface Committer {
	Name?: string;
	Email?: string;
	Commits?: number;
}

export interface CommitterList {
	Committers?: any[];
	HasMore?: boolean;
}

export interface CommittersOptions {
}

export interface CompletionItem {
	label: string;
	kind?: number;
	detail?: string;
	documentation?: string;
	sortText?: string;
	filterText?: string;
	insertText?: string;
	textEdit?: any;
	data?: any;
}

export interface CompletionList {
	isIncomplete: boolean;
	items: any[];
}

export interface CompletionOptions {
	resolveProvider?: boolean;
	triggerCharacters?: string[];
}

export interface Def {
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path: string;
	Name: string;
	Kind?: string;
	File: string;
	DefStart: number;
	DefEnd: number;
	Exported?: boolean;
	Local?: boolean;
	Test?: boolean;
	Data?: any;
	Docs?: any[];
	TreePath?: string;
	DocHTML?: any;
	FmtStrings?: any;
	StartLine?: number;
	EndLine?: number;
}

export interface DefAuthor {
	Email?: string;
	AvatarURL?: string;
	LastCommitDate: any;
	LastCommitID?: string;
	Bytes?: number;
	BytesProportion?: number;
}

export interface DefAuthorList {
	DefAuthors?: DefAuthor[];
}

export interface DefAuthorship {
	LastCommitDate: any;
	LastCommitID?: string;
	Bytes?: number;
	BytesProportion?: number;
}

export interface DefDoc {
	Format: string;
	Data: string;
}

export interface DefFileRef {
	Path?: string;
	Count?: number;
	Score?: number;
}

export interface DefFormatStrings {
	Name: any;
	Type: any;
	NameAndTypeSeparator?: string;
	Language?: string;
	DefKeyword?: string;
	Kind?: string;
}

export interface DefGetOptions {
	Doc?: boolean;
	ComputeLineRange?: boolean;
}

export interface DefKey {
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path: string;
}

export interface DefList {
	Defs?: Def[];
	Total?: number;
}

export interface DefListAuthorsOptions {
	PerPage?: number;
	Page?: number;
}

export interface DefListOptions {
	Name?: string;
	Query?: string;
	ByteStart?: number;
	ByteEnd?: number;
	DefKeys?: any[];
	RepoRevs?: string[];
	UnitType?: string;
	Unit?: string;
	Path?: string;
	Files?: string[];
	FilePathPrefix?: string;
	Kinds?: string[];
	Exported?: boolean;
	Nonlocal?: boolean;
	IncludeTest?: boolean;
	Doc?: boolean;
	Fuzzy?: boolean;
	Sort?: string;
	Direction?: string;
	PerPage?: number;
	Page?: number;
}

export interface DefListRefLocationsOptions {
	Repos?: string[];
	PerPage?: number;
	Page?: number;
}

export interface DefListRefsOptions {
	Repo?: number;
	CommitID?: string;
	Files?: string[];
	PerPage?: number;
	Page?: number;
}

export interface DefRepoRef {
	Repo?: string;
	Count?: number;
	Score?: number;
	Files?: DefFileRef[];
}

export interface DefSearchResult {
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path: string;
	Name: string;
	Kind?: string;
	File: string;
	DefStart: number;
	DefEnd: number;
	Exported?: boolean;
	Local?: boolean;
	Test?: boolean;
	Data?: any;
	Docs?: any[];
	TreePath?: string;
	DocHTML?: any;
	FmtStrings?: any;
	StartLine?: number;
	EndLine?: number;
	Score?: number;
	RefCount?: number;
}

export interface DefSpec {
	Repo?: number;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path?: string;
}

export interface DefsGetOp {
	Def: DefSpec;
	Opt?: DefGetOptions;
}

export interface DefsListAuthorsOp {
	Def: DefSpec;
	Opt?: DefListAuthorsOptions;
}

export interface DefsListRefLocationsOp {
	Def: DefSpec;
	Opt?: DefListRefLocationsOptions;
}

export interface DefsListRefsOp {
	Def: DefSpec;
	Opt?: DefListRefsOptions;
}

export interface DefsRefreshIndexOp {
	Repo?: number;
	RefreshRefLocations?: boolean;
	Force?: boolean;
}

export interface Diagnostic {
	range: any;
	severity?: any;
	code?: string;
	source?: string;
	message: string;
}

export interface DidChangeConfigurationParams {
	settings: any;
}

export interface DidChangeTextDocumentParams {
	textDocument: any;
	contentChanges: any[];
}

export interface DidChangeWatchedFilesParams {
	changes: any[];
}

export interface DidCloseTextDocumentParams {
	textDocument: any;
}

export interface DidOpenTextDocumentParams {
	textDocument: any;
}

export interface DidSaveTextDocumentParams {
	textDocument: any;
}

export interface Diff {
}

export interface DiffOptions {
}

export interface Doc {
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path: string;
	Format: string;
	Data: string;
	File?: string;
	Start?: number;
	End?: number;
	DocUnit?: string;
}

export interface DocKey {
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Path: string;
}

export interface DocumentFormattingParams {
	textDocument: any;
	options: any;
}

export interface DocumentHighlight {
	range: any;
	kind?: number;
}

export interface DocumentOnTypeFormattingOptions {
	firstTriggerCharacter: string;
	moreTriggerCharacter?: string[];
}

export interface DocumentOnTypeFormattingParams {
	textDocument: any;
	position: any;
	ch: string;
	formattingOptions: any;
}

export interface DocumentRangeFormattingParams {
	textDocument: any;
	range: any;
	options: any;
}

export interface DocumentSymbolParams {
	textDocument: any;
}

export interface EmailAddr {
	Email?: string;
	Verified?: boolean;
	Primary?: boolean;
	Guessed?: boolean;
	Blacklisted?: boolean;
}

export interface EmailAddrList {
	EmailAddrs?: EmailAddr[];
}

export interface Event {
	Type?: string;
	UserID?: string;
	DeviceID?: string;
	Timestamp?: any;
	UserProperties?: any;
	EventProperties?: any;
}

export interface EventList {
	Events?: Event[];
	Version?: string;
	AppURL?: string;
}

export interface ExternalToken {
	uid?: string;
	host?: string;
	token?: string;
	scope?: string;
}

export interface FileData {
}

export interface FileEvent {
	uri: string;
	type: number;
}

export interface FileRange {
	StartLine?: number;
	EndLine?: number;
	StartByte?: number;
	EndByte?: number;
}

export interface FileWithRange {
	StartLine?: number;
	EndLine?: number;
	StartByte?: number;
	EndByte?: number;
}

export interface FormattingOptions {
	tabSize: number;
	insertSpaces: boolean;
	key: string;
}

export interface GRPCCodec {
}

export interface GetFileOptions {
	StartLine?: number;
	EndLine?: number;
	StartByte?: number;
	EndByte?: number;
	EntireFile?: boolean;
	ExpandContextLines?: number;
	FullLines?: boolean;
	Recursive?: boolean;
	RecurseSingleSubfolderLimit?: number;
}

export interface HTTPSConfig {
}

export interface Hover {
	contents?: any[];
	range: any;
}

export interface Hunk {
}

export interface InitializeError {
	retry: boolean;
}

export interface InitializeParams {
	processId?: number;
	rootPath: string;
	capabilities?: any;
}

export interface InitializeResult {
	capabilities?: any;
}

export interface ListOptions {
	PerPage?: number;
	Page?: number;
}

export interface ListResponse {
	Total?: number;
}

export interface Location {
	uri: string;
	range: any;
}

export interface LogMessageParams {
	type: number;
	message: string;
}

export interface MarkedString {
	language: string;
	value: string;
}

export interface MessageActionItem {
	title: string;
}

export interface MirrorReposRefreshVCSOp {
	Repo?: number;
	AsUser?: UserSpec;
}

export interface None {
}

export interface Origin {
	ID?: string;
	APIBaseURL?: string;
}

export interface Output {
	Defs?: any[];
	Refs?: any[];
	Docs?: any[];
	Anns?: any[];
}

export interface Packet {
	data?: number[];
}

export interface ParameterInformation {
	label: string;
	documentation?: string;
}

export interface Position {
	line: number;
	character: number;
}

export interface Propagate {
}

export interface PublishDiagnosticsParams {
	uri: string;
	diagnostics: any[];
}

export interface QualFormatStrings {
	Unqualified?: string;
	ScopeQualified?: string;
	DepQualified?: string;
	RepositoryWideQualified?: string;
	LanguageWideQualified?: string;
}

export interface Range {
}

export interface ReceivePackOp {
	repo?: number;
	data?: number[];
	advertiseRefs?: boolean;
}

export interface Ref {
	DefRepo?: string;
	DefUnitType?: string;
	DefUnit?: string;
	DefPath: string;
	Repo?: string;
	CommitID?: string;
	UnitType?: string;
	Unit?: string;
	Def?: boolean;
	File?: string;
	Start: number;
	End: number;
}

export interface RefDefKey {
	DefRepo?: string;
	DefUnitType?: string;
	DefUnit?: string;
	DefPath: string;
}

export interface RefKey {
}

export interface RefList {
	Refs?: any[];
	HasMore?: boolean;
}

export interface RefLocationsList {
	RepoRefs?: DefRepoRef[];
	HasMore?: boolean;
	TotalRepos?: number;
}

export interface RefSet {
}

export interface ReferenceContext {
	IncludeDeclaration: boolean;
}

export interface ReferenceParams {
	textDocument: any;
	position: any;
	context: any;
}

export interface RemoteOpts {
}

export interface RemoteRepo {
	GitHubID?: number;
	Owner?: string;
	OwnerIsOrg?: boolean;
	Name?: string;
	VCS?: string;
	HTTPCloneURL?: string;
	DefaultBranch?: string;
	Description?: string;
	Language?: string;
	UpdatedAt?: any;
	PushedAt?: any;
	Private?: boolean;
	Fork?: boolean;
	Mirror?: boolean;
	Stars?: number;
	Permissions?: RepoPermissions;
}

export interface RenameParams {
	textDocument: any;
	position: any;
	newName: string;
}

export interface Repo {
	ID?: number;
	URI?: string;
	Owner?: string;
	Name?: string;
	Description?: string;
	HTTPCloneURL?: string;
	SSHCloneURL?: string;
	HomepageURL?: string;
	DefaultBranch?: string;
	Language?: string;
	Blocked?: boolean;
	Deprecated?: boolean;
	Fork?: boolean;
	Mirror?: boolean;
	Private?: boolean;
	CreatedAt?: any;
	UpdatedAt?: any;
	PushedAt?: any;
	VCSSyncedAt?: any;
	Origin?: Origin;
	Permissions?: RepoPermissions;
}

export interface RepoConfig {
}

export interface RepoList {
	Repos?: Repo[];
}

export interface RepoListBranchesOptions {
	IncludeCommit?: boolean;
	BehindAheadBranch?: string;
	ContainsCommit?: string;
	PerPage?: number;
	Page?: number;
}

export interface RepoListCommitsOptions {
	Head?: string;
	Base?: string;
	PerPage?: number;
	Page?: number;
	Path?: string;
}

export interface RepoListCommittersOptions {
	Rev?: string;
	PerPage?: number;
	Page?: number;
}

export interface RepoListOptions {
	Name?: string;
	Query?: string;
	URIs?: string[];
	Sort?: string;
	Direction?: string;
	NoFork?: boolean;
	Type?: string;
	Owner?: string;
	RemoteOnly?: boolean;
	RemoteSearch?: boolean;
	PerPage?: number;
	Page?: number;
}

export interface RepoListTagsOptions {
	PerPage?: number;
	Page?: number;
}

export interface RepoNotExistError {
}

export interface RepoPermissions {
	Pull?: boolean;
	Push?: boolean;
	Admin?: boolean;
}

export interface RepoResolution {
	Repo?: number;
	CanonicalPath?: string;
	RemoteRepo?: Repo;
}

export interface RepoResolveOp {
	path?: string;
	remote?: boolean;
}

export interface RepoRevSpec {
	Repo?: number;
	CommitID?: string;
}

export interface RepoSearchResult {
	Repo?: Repo;
}

export interface RepoSpec {
	ID?: number;
}

export interface RepoStatus {
	State?: string;
	TargetURL?: string;
	Description?: string;
	Context?: string;
	CreatedAt: any;
	UpdatedAt: any;
}

export interface RepoStatusList {
	RepoStatuses?: RepoStatus[];
}

export interface RepoStatusesCreateOp {
	Repo: RepoRevSpec;
	Status: RepoStatus;
}

export interface RepoTreeGetOp {
	Entry: TreeEntrySpec;
	Opt?: RepoTreeGetOptions;
}

export interface RepoTreeGetOptions {
	ContentsAsString?: boolean;
	StartLine?: number;
	EndLine?: number;
	StartByte?: number;
	EndByte?: number;
	EntireFile?: boolean;
	ExpandContextLines?: number;
	FullLines?: boolean;
	Recursive?: boolean;
	RecurseSingleSubfolderLimit?: number;
	NoSrclibAnns?: boolean;
}

export interface RepoTreeListOp {
	Rev: RepoRevSpec;
}

export interface RepoTreeListResult {
	Files?: string[];
}

export interface RepoWebhookOptions {
	URI?: string;
}

export interface ReposCreateOp {
}

export interface ReposListBranchesOp {
	Repo?: number;
	Opt?: RepoListBranchesOptions;
}

export interface ReposListCommitsOp {
	Repo?: number;
	Opt?: RepoListCommitsOptions;
}

export interface ReposListCommittersOp {
	Repo?: number;
	Opt?: RepoListCommittersOptions;
}

export interface ReposListTagsOp {
	Repo?: number;
	Opt?: RepoListTagsOptions;
}

export interface ReposResolveRevOp {
	repo?: number;
	rev?: string;
}

export interface ReposUpdateOp {
	Repo?: number;
	URI?: string;
	Owner?: string;
	Name?: string;
	Description?: string;
	HTTPCloneURL?: string;
	SSHCloneURL?: string;
	HomepageURL?: string;
	DefaultBranch?: string;
	Language?: string;
	Origin?: Origin;
	Blocked?: any;
	Deprecated?: any;
	Fork?: any;
	Mirror?: any;
	Private?: any;
}

export interface RepositoryListingDef {
}

export interface ResolvedRev {
	CommitID?: string;
}

export interface ResponseError {
}

export interface SSHConfig {
}

export interface SearchOp {
	Query?: string;
	Opt?: SearchOptions;
}

export interface SearchOptions {
	Repos?: number[];
	NotRepos?: number[];
	Languages?: string[];
	NotLanguages?: string[];
	Kinds?: string[];
	NotKinds?: string[];
	PerPage?: number;
	Page?: number;
	IncludeRepos?: boolean;
	Fast?: boolean;
	AllowEmpty?: boolean;
}

export interface SearchReposResultList {
	Repos?: Repo[];
}

export interface SearchResult {
	File?: string;
	StartByte?: number;
	EndByte?: number;
	StartLine?: number;
	EndLine?: number;
	Match?: number[];
}

export interface SearchResultsList {
	RepoResults?: RepoSearchResult[];
	DefResults?: DefSearchResult[];
	SearchQueryOptions?: SearchOptions[];
}

export interface ServerCapabilities {
	textDocumentSync?: number;
	hoverProvider?: boolean;
	completionProvider?: any;
	signatureHelpProvider?: any;
	definitionProvider?: boolean;
	referencesProvider?: boolean;
	documentHighlightProvider?: boolean;
	documentSymbolProvider?: boolean;
	workspaceSymbolProvider?: boolean;
	codeActionProvider?: boolean;
	codeLensProvider?: any;
	documentFormattingProvider?: boolean;
	documentRangeFormattingProvider?: boolean;
	documentOnTypeFormattingProvider?: any;
	renameProvider?: boolean;
}

export interface ServerConfig {
	Version?: string;
	AppURL?: string;
}

export interface ShowMessageParams {
	type: number;
	message: string;
}

export interface ShowMessageRequestParams {
	type: number;
	message: string;
	actions: any[];
}

export interface Signature {
	Name?: string;
	Email?: string;
	Date: any;
}

export interface SignatureHelp {
	signatures: any[];
	activeSignature?: number;
	activeParameter?: number;
}

export interface SignatureHelpOptions {
	triggerCharacters?: string[];
}

export interface SignatureInformation {
	label: string;
	documentation?: string;
	paramaters?: any[];
}

export interface SrclibDataVersion {
	CommitID?: string;
	CommitsBehind?: number;
}

export interface StreamResponse {
	HasMore?: boolean;
}

export interface SubmoduleInfo {
}

export interface SymbolInformation {
	name: string;
	kind: any;
	location: any;
	containerName?: string;
}

export interface Tag {
	Name?: string;
	CommitID?: any;
}

export interface TagList {
	Tags?: any[];
	HasMore?: boolean;
}

export interface TextDocumentContentChangeEvent {
	range: any;
	rangeLength: number;
	text: string;
}

export interface TextDocumentIdentifier {
	uri: string;
}

export interface TextDocumentItem {
	uri: string;
	languageId: string;
	version: number;
	text: string;
}

export interface TextDocumentPositionParams {
	textDocument: any;
	position: any;
}

export interface TextEdit {
	range: any;
	newText: string;
}

export interface TreeEntry {
	ContentsString?: string;
}

export interface TreeEntrySpec {
	RepoRev: RepoRevSpec;
	Path?: string;
}

export interface URIList {
	URIs?: string[];
}

export interface UpdateEmailsOp {
	UserSpec: UserSpec;
	Add?: EmailAddrList;
}

export interface UpdateResult {
}

export interface UploadPackOp {
	repo?: number;
	data?: number[];
	advertiseRefs?: boolean;
}

export interface User {
	UID: string;
	Login: string;
	Name?: string;
	IsOrganization?: boolean;
	AvatarURL?: string;
	Location?: string;
	Company?: string;
	HomepageURL?: string;
	Disabled?: boolean;
	Admin?: boolean;
	Betas?: string[];
	Write?: boolean;
	RegisteredAt?: any;
}

export interface UserEvent {
	Type?: string;
	UID?: string;
	Service?: string;
	Method?: string;
	Result?: string;
	CreatedAt?: any;
	Message?: string;
	Version?: string;
	URL?: string;
}

export interface UserList {
	Users?: User[];
}

export interface UserSpec {
	UID?: string;
}

export interface VCSSearchResultList {
	SearchResults?: any[];
	Total?: number;
}

export interface VersionedTextDocumentIdentifier {
	uri: string;
	version: number;
}

export interface WorkspaceEdit {
	changes: any;
}

export interface WorkspaceSymbolParams {
	query: string;
}
