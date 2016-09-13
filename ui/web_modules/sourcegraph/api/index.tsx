// GENERATED CODE - DO NOT EDIT!

export interface AccessTokenRequest {
	Scope?: string[];
}

export interface AccessTokenResponse {
	AccessToken?: string;
	Scope?: string[];
	UID?: number;
	GitHubAccessToken?: string;
	GitHubUser?: GitHubUser;
}

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
	UID?: number;
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

export interface Build {
	Repo?: number;
	ID?: number;
	CommitID?: string;
	Branch?: string;
	Tag?: string;
	CreatedAt: any;
	StartedAt?: any;
	EndedAt?: any;
	HeartbeatAt?: any;
	Success?: boolean;
	Failure?: boolean;
	Killed?: boolean;
	Host?: string;
	Purged?: boolean;
	Queue?: boolean;
	Priority?: number;
	BuilderConfig?: string;
}

export interface BuildConfig {
	Queue?: boolean;
	Priority?: number;
	BuilderConfig?: string;
}

export interface BuildGetLogOptions {
	MinID?: string;
}

export interface BuildJob {
	Spec: BuildSpec;
	CommitID?: string;
	Branch?: string;
	Tag?: string;
	AccessToken?: string;
}

export interface BuildList {
	Builds?: Build[];
	HasMore?: boolean;
}

export interface BuildListOptions {
	Queued?: boolean;
	Active?: boolean;
	Ended?: boolean;
	Succeeded?: boolean;
	Failed?: boolean;
	Purged?: boolean;
	Repo?: number;
	CommitID?: string;
	Sort?: string;
	Direction?: string;
	PerPage?: number;
	Page?: number;
}

export interface BuildSpec {
	Repo?: number;
	ID?: number;
}

export interface BuildTask {
	ID?: number;
	Build: BuildSpec;
	ParentID?: number;
	Label?: string;
	CreatedAt: any;
	StartedAt?: any;
	EndedAt?: any;
	Success?: boolean;
	Failure?: boolean;
	Skipped?: boolean;
	Warnings?: boolean;
}

export interface BuildTaskList {
	BuildTasks?: BuildTask[];
}

export interface BuildTaskListOptions {
	PerPage?: number;
	Page?: number;
}

export interface BuildUpdate {
	StartedAt?: any;
	EndedAt?: any;
	HeartbeatAt?: any;
	Host?: string;
	Success?: boolean;
	Purged?: boolean;
	Failure?: boolean;
	Killed?: boolean;
	Priority?: number;
	BuilderConfig?: string;
	FileScore?: number;
	RefScore?: number;
	TokDensity?: number;
}

export interface BuildsCreateOp {
	Repo?: number;
	CommitID?: string;
	Branch?: string;
	Tag?: string;
	Config: BuildConfig;
}

export interface BuildsCreateTasksOp {
	Build: BuildSpec;
	Tasks?: BuildTask[];
}

export interface BuildsDequeueNextOp {
}

export interface BuildsGetTaskLogOp {
	Task: TaskSpec;
	Opt?: BuildGetLogOptions;
}

export interface BuildsListBuildTasksOp {
	Build: BuildSpec;
	Opt?: BuildTaskListOptions;
}

export interface BuildsUpdateOp {
	Build: BuildSpec;
	Info: BuildUpdate;
}

export interface BuildsUpdateTaskOp {
	Task: TaskSpec;
	Info: TaskUpdate;
}

export interface Change {
}

export interface Client {
}

export interface CombinedStatus {
	Rev?: string;
	CommitID?: string;
	State?: string;
	Statuses?: RepoStatus[];
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

export interface CreatedAccount {
	UID?: number;
	TemporaryAccessToken?: string;
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

export interface DefsListExamplesOp {
	Def: DefSpec;
	PerPage?: number;
	Page?: number;
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
	ClientID?: string;
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
	uid?: number;
	host?: string;
	token?: string;
	scope?: string;
	client_id?: string;
	ext_uid?: number;
}

export interface ExternalTokenSpec {
	uid?: number;
	host?: string;
	client_id?: string;
}

export interface FileData {
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

export interface GitHubAuthCode {
	Code?: string;
	Host?: string;
}

export interface GitHubUser {
	ID?: number;
	Login?: string;
	Name?: string;
	Email?: string;
	Location?: string;
	Company?: string;
	AvatarURL?: string;
}

export interface HTTPSConfig {
}

export interface Hunk {
}

export interface ListOptions {
	PerPage?: number;
	Page?: number;
}

export interface ListResponse {
	Total?: number;
}

export interface LogEntries {
	MaxID?: string;
	Entries?: string[];
}

export interface LoginCredentials {
	Login?: string;
	Password?: string;
}

export interface MirrorReposRefreshVCSOp {
	Repo?: number;
	AsUser?: UserSpec;
}

export interface NewAccount {
	Login?: string;
	Email?: string;
	Password?: string;
	UID?: number;
}

export interface NewPassword {
	Password?: string;
	Token?: PasswordResetToken;
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

export interface PasswordResetToken {
	Token?: string;
}

export interface PendingPasswordReset {
	Link?: string;
	Token?: PasswordResetToken;
	EmailSent?: boolean;
	Login?: string;
}

export interface Propagate {
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

export interface Repo {
	ID?: number;
	URI?: string;
	Owner?: string;
	Name?: string;
	Description?: string;
	HTTPCloneURL?: string;
	SSHCloneURL?: string;
	HomepageURL?: string;
	HTMLURL?: string;
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

export interface RepoTreeSearchOp {
	Rev: RepoRevSpec;
	Opt?: RepoTreeSearchOptions;
}

export interface RepoTreeSearchOptions {
	Query?: string;
	QueryType?: string;
	ContextLines?: number;
	N?: number;
	Offset?: number;
}

export interface RepoTreeSearchResult {
	File?: string;
	StartByte?: number;
	EndByte?: number;
	StartLine?: number;
	EndLine?: number;
	Match?: number[];
	RepoRev: RepoRevSpec;
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
	Description?: string;
	Language?: string;
	DefaultBranch?: string;
	Fork?: any;
	Private?: any;
}

export interface RepositoryListingDef {
}

export interface RequestPasswordResetOp {
	Email?: string;
}

export interface ResolvedRev {
	CommitID?: string;
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

export interface SearchReposOp {
	Query?: string;
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

export interface ServerConfig {
	Version?: string;
	AppURL?: string;
	IDKey?: string;
}

export interface Signature {
	Name?: string;
	Email?: string;
	Date: any;
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

export interface Tag {
	Name?: string;
	CommitID?: any;
}

export interface TagList {
	Tags?: any[];
	HasMore?: boolean;
}

export interface TaskSpec {
	Build: BuildSpec;
	ID?: number;
}

export interface TaskUpdate {
	StartedAt?: any;
	EndedAt?: any;
	Success?: boolean;
	Failure?: boolean;
	Skipped?: boolean;
	Warnings?: boolean;
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
	UID: number;
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
	BetaRegistered?: boolean;
	Write?: boolean;
	RegisteredAt?: any;
}

export interface UserEvent {
	Type?: string;
	UID?: number;
	ClientID?: string;
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
	Login?: string;
	UID?: number;
}

export interface UsersListOptions {
	Query?: string;
	Sort?: string;
	Direction?: string;
	PerPage?: number;
	Page?: number;
	UIDs?: number[];
	AllBetas?: string[];
	RegisteredBeta?: boolean;
	HaveBeta?: boolean;
}

export interface VCSCredentials {
	Pass?: string;
}

export interface VCSSearchResultList {
	SearchResults?: any[];
	Total?: number;
}
