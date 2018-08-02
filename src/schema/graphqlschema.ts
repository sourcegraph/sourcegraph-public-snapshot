export type ID = string

export interface IGraphQLResponseRoot {
    data?: IQuery | IMutation
    errors?: IGraphQLResponseError[]
}

export interface IGraphQLResponseError {
    /** Required for all errors */
    message: string
    locations?: IGraphQLResponseErrorLocation[]
    /** 7.2.2 says 'GraphQL servers may provide additional entries to error' */
    [propName: string]: any
}

export interface IGraphQLResponseErrorLocation {
    line: number
    column: number
}

/**
 * A query.
 */
export interface IQuery {
    __typename: 'Query'

    /**
     * The root of the query.
     * @deprecated this will be removed.
     */
    root: IQuery

    /**
     * Looks up a node by ID.
     */
    node: Node | null

    /**
     * Looks up a repository by name.
     */
    repository: IRepository | null

    /**
     * List all repositories.
     */
    repositories: IRepositoryConnection

    /**
     * Looks up a Phabricator repository by name.
     */
    phabricatorRepo: IPhabricatorRepo | null

    /**
     * The current user.
     */
    currentUser: IUser | null

    /**
     * Looks up a user by username.
     */
    user: IUser | null

    /**
     * List all users.
     */
    users: IUserConnection

    /**
     * Looks up an organization by name.
     */
    organization: IOrg | null

    /**
     * List all organizations.
     */
    organizations: IOrgConnection

    /**
     * Lists discussion threads.
     */
    discussionThreads: IDiscussionThreadConnection

    /**
     * Lists discussion comments.
     */
    discussionComments: IDiscussionCommentConnection

    /**
     * Renders Markdown to HTML. The returned HTML is already sanitized and
     * escaped and thus is always safe to render.
     */
    renderMarkdown: string

    /**
     * Looks up an instance of a type that implements ConfigurationSubject.
     */
    configurationSubject: ConfigurationSubject | null

    /**
     * The configuration for the viewer.
     */
    viewerConfiguration: IConfigurationCascade

    /**
     * Runs a search.
     */
    search: ISearch | null

    /**
     * The search scopes.
     */
    searchScopes: ISearchScope[]

    /**
     * All saved queries configured for the current user, merged from all configurations.
     */
    savedQueries: ISavedQuery[]

    /**
     * All repository groups for the current user, merged from all configurations.
     */
    repoGroups: IRepoGroup[]

    /**
     * The current site.
     */
    site: ISite

    /**
     * Retrieve responses to surveys.
     */
    surveyResponses: ISurveyResponseConnection

    /**
     * The extension registry.
     */
    extensionRegistry: IExtensionRegistry
}

export interface INodeOnQueryArguments {
    id: ID
}

export interface IRepositoryOnQueryArguments {
    /**
     * The name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null
}

export interface IRepositoriesOnQueryArguments {
    /**
     * Returns the first n repositories from the list.
     */
    first?: number | null

    /**
     * Return repositories whose names match the query.
     */
    query?: string | null

    /**
     * Include enabled repositories.
     * @default true
     */
    enabled?: boolean | null

    /**
     * Include disabled repositories.
     * @default false
     */
    disabled?: boolean | null

    /**
     * Include cloned repositories.
     * @default true
     */
    cloned?: boolean | null

    /**
     * Include repositories that are currently being cloned.
     * @default true
     */
    cloneInProgress?: boolean | null

    /**
     * Include repositories that are not yet cloned and for which cloning is not in progress.
     * @default true
     */
    notCloned?: boolean | null

    /**
     * Include repositories that have a text search index.
     * @default true
     */
    indexed?: boolean | null

    /**
     * Include repositories that do not have a text search index.
     * @default true
     */
    notIndexed?: boolean | null

    /**
     * Filter for repositories that have been indexed for cross-repository code intelligence.
     * @default false
     */
    ciIndexed?: boolean | null

    /**
     * Filter for repositories that have not been indexed for cross-repository code intelligence.
     * @default false
     */
    notCIIndexed?: boolean | null

    /**
     * Sort field.
     * @default REPO_URI
     */
    orderBy?: RepoOrderBy | null

    /**
     * Sort direction.
     * @default false
     */
    descending?: boolean | null
}

export interface IPhabricatorRepoOnQueryArguments {
    /**
     * The name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null
}

export interface IUserOnQueryArguments {
    username: string
}

export interface IUsersOnQueryArguments {
    /**
     * Returns the first n users from the list.
     */
    first?: number | null

    /**
     * Return users whose usernames or display names match the query.
     */
    query?: string | null

    /**
     * Returns users who have been active in a given period of time.
     */
    activePeriod?: UserActivePeriod | null
}

export interface IOrganizationOnQueryArguments {
    name: string
}

export interface IOrganizationsOnQueryArguments {
    /**
     * Returns the first n organizations from the list.
     */
    first?: number | null

    /**
     * Return organizations whose names or display names match the query.
     */
    query?: string | null
}

export interface IDiscussionThreadsOnQueryArguments {
    /**
     * Returns the first n threads from the list.
     */
    first?: number | null

    /**
     * When present, lists only the thread with this ID.
     */
    threadID?: ID | null

    /**
     * When present, lists only the threads created by this author.
     */
    authorUserID?: ID | null

    /**
     * When present, lists only the threads whose target is a repository with this ID.
     */
    targetRepositoryID?: ID | null

    /**
     * When present, lists only the threads whose target is a repository with this file path.
     */
    targetRepositoryPath?: string | null
}

export interface IDiscussionCommentsOnQueryArguments {
    /**
     * Returns the first n comments from the list.
     */
    first?: number | null

    /**
     * When present, lists only the comments created by this author.
     */
    authorUserID?: ID | null
}

export interface IRenderMarkdownOnQueryArguments {
    markdown: string
    options?: IMarkdownOptions | null
}

export interface IConfigurationSubjectOnQueryArguments {
    id: ID
}

export interface ISearchOnQueryArguments {
    /**
     * The search query (such as "foo" or "repo:myrepo foo").
     * @default
     */
    query?: string | null
}

export interface ISurveyResponsesOnQueryArguments {
    /**
     * Returns the first n survey responses from the list.
     */
    first?: number | null
}

/**
 * An object with an ID.
 */
export type Node =
    | IRepository
    | IGitCommit
    | IUser
    | IOrg
    | IOrganizationInvitation
    | IRegistryExtension
    | IAccessToken
    | IExternalAccount
    | IPackage
    | IDependency
    | IGitRef

/**
 * An object with an ID.
 */
export interface INode {
    __typename: 'Node'

    /**
     * The ID of the node.
     */
    id: ID
}

/**
 * A repository is a Git source control repository that is mirrored from some origin code host.
 */
export interface IRepository {
    __typename: 'Repository'

    /**
     * The repository's unique ID.
     */
    id: ID

    /**
     *  The repository's name, as a path with one or more components. It conventionally consists of
     *  the repository's hostname and path (joined by "/"), minus any suffixes (such as ".git").
     *
     *  Examples:
     *
     *  - github.com/foo/bar
     *  - my-code-host.example.com/myrepo
     *  - myrepo
     */
    name: string

    /**
     * An alias for name.
     * @deprecated use name instead
     */
    uri: string

    /**
     * The repository's description.
     */
    description: string

    /**
     * The primary programming language in the repository.
     */
    language: string

    /**
     *  Whether the repository is enabled. A disabled repository should only be accessible to site admins.
     *
     *  NOTE: Disabling a repository does not provide any additional security. This field is merely a
     *  guideline to UI implementations.
     */
    enabled: boolean

    /**
     * The date when this repository was created on Sourcegraph.
     */
    createdAt: string

    /**
     * The date when this repository's metadata was last updated on Sourcegraph.
     */
    updatedAt: string | null

    /**
     * Returns information about the given commit in the repository, or null if no commit exists with the given rev.
     */
    commit: IGitCommit | null

    /**
     * Information and status related to mirroring, if this repository is a mirror of another repository (e.g., on
     * some code host). In this case, the remote source repository is external to Sourcegraph and the mirror is
     * maintained by the Sourcegraph site (not the other way around).
     */
    mirrorInfo: IMirrorRepositoryInfo

    /**
     * Information about this repository from the external service that it originates from (such as GitHub, GitLab,
     * Phabricator, etc.).
     */
    externalRepository: IExternalRepository | null

    /**
     * Whether the repository is currently being cloned.
     * @deprecated use Repository.mirrorInfo.cloneInProgress instead
     */
    cloneInProgress: boolean

    /**
     * The commit that was last indexed for cross-references, if any.
     */
    lastIndexedRevOrLatest: IGitCommit | null

    /**
     * Information about the text search index for this repository, or null if text search indexing
     * is not enabled or supported for this repository.
     */
    textSearchIndex: IRepositoryTextSearchIndex | null

    /**
     * The URL to this repository.
     */
    url: string

    /**
     * The URLs to this repository on external services associated with it.
     */
    externalURLs: IExternalLink[]

    /**
     * The repository's default Git branch (HEAD symbolic ref). If the repository is currently being cloned or is
     * empty, this field will be null.
     */
    defaultBranch: IGitRef | null

    /**
     * The repository's Git refs.
     */
    gitRefs: IGitRefConnection

    /**
     * The repository's Git branches.
     */
    branches: IGitRefConnection

    /**
     * The repository's Git tags.
     */
    tags: IGitRefConnection

    /**
     * A Git comparison in this repository between a base and head commit.
     */
    comparison: IRepositoryComparison

    /**
     * The repository's contributors.
     */
    contributors: IRepositoryContributorConnection

    /**
     *  The repository's symbols (e.g., functions, variables, types, classes, etc.) on the default branch.
     *
     *  The result may be stale if a new commit was just pushed to this repository's default branch and it has not
     *  yet been processed. Use Repository.commit.tree.symbols to retrieve symbols for a specific revision.
     */
    symbols: ISymbolConnection

    /**
     *  Packages defined in this repository, as returned by LSP workspace/xpackages requests to this repository's
     *  language servers (running against a recent commit on its default branch).
     *
     *  The result may be stale if a new commit was just pushed to this repository's default branch and it has not
     *  yet been processed. Use Repository.commit.packages to retrieve packages for a specific revision.
     */
    packages: IPackageConnection

    /**
     *  Dependencies of this repository, as returned by LSP workspace/xreferences requests to this repository's
     *  language servers (running against a recent commit on its default branch).
     *
     *  The result may be stale if a new commit was just pushed to this repository's default branch and it has not
     *  yet been processed. Use Repository.commit.dependencies to retrieve dependencies for a specific revision.
     */
    dependencies: IDependencyConnection

    /**
     * The total ref list.
     */
    listTotalRefs: ITotalRefList

    /**
     * Link to another Sourcegraph instance location where this repository is located.
     */
    redirectURL: string | null

    /**
     * Whether the viewer has admin privileges on this repository.
     */
    viewerCanAdminister: boolean
}

export interface ICommitOnRepositoryArguments {
    /**
     * The Git revision specifier (revspec) for the commit.
     */
    rev: string

    /**
     * Optional input revspec used to construct non-canonical URLs and other "friendly" field values. Used by
     * clients that must ensure consistency of revision resolution within a session/request (so they use full
     * SHAs) but also preserve the user input rev (for user friendliness).
     */
    inputRevspec?: string | null
}

export interface IGitRefsOnRepositoryArguments {
    /**
     * Returns the first n Git refs from the list.
     */
    first?: number | null

    /**
     * Return Git refs whose names match the query.
     */
    query?: string | null

    /**
     *  Return only Git refs of the given type.
     *
     *  Known issue: It is only supported to retrieve Git branch and tag refs, not
     *  other Git refs.
     */
    type?: GitRefType | null

    /**
     * Ordering for Git refs in the list.
     */
    orderBy?: GitRefOrder | null
}

export interface IBranchesOnRepositoryArguments {
    /**
     * Returns the first n Git branches from the list.
     */
    first?: number | null

    /**
     * Return Git branches whose names match the query.
     */
    query?: string | null

    /**
     * Ordering for Git branches in the list.
     */
    orderBy?: GitRefOrder | null
}

export interface ITagsOnRepositoryArguments {
    /**
     * Returns the first n Git tags from the list.
     */
    first?: number | null

    /**
     * Return Git tags whose names match the query.
     */
    query?: string | null
}

export interface IComparisonOnRepositoryArguments {
    /**
     * The base of the diff ("old" or "left-hand side"), or "HEAD" if not specified.
     */
    base?: string | null

    /**
     * The head of the diff ("new" or "right-hand side"), or "HEAD" if not specified.
     */
    head?: string | null
}

export interface IContributorsOnRepositoryArguments {
    /**
     * The Git revision range to compute contributors in.
     */
    revisionRange?: string | null

    /**
     * The date after which to count contributions.
     */
    after?: string | null

    /**
     * Return contributors to files in this path.
     */
    path?: string | null

    /**
     * Returns the first n contributors from the list.
     */
    first?: number | null
}

export interface ISymbolsOnRepositoryArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

export interface IPackagesOnRepositoryArguments {
    /**
     * Returns the first n packages from the list.
     */
    first?: number | null

    /**
     * Return packages matching the query.
     */
    query?: string | null
}

export interface IDependenciesOnRepositoryArguments {
    /**
     * Returns the first n dependencies from the list.
     */
    first?: number | null

    /**
     * Return dependencies matching the query.
     */
    query?: string | null
}

/**
 * A Git commit.
 */
export interface IGitCommit {
    __typename: 'GitCommit'

    /**
     * The globally addressable ID for this commit.
     */
    id: ID

    /**
     * The repository that contains this commit.
     */
    repository: IRepository

    /**
     * This commit's Git object ID (OID), a 40-character SHA-1 hash.
     */
    oid: any

    /**
     * The abbreviated form of this commit's OID.
     */
    abbreviatedOID: string

    /**
     * This commit's author.
     */
    author: ISignature

    /**
     * This commit's committer, if any.
     */
    committer: ISignature | null

    /**
     * The full commit message.
     */
    message: string

    /**
     * The first line of the commit message.
     */
    subject: string

    /**
     * The contents of the commit message after the first line.
     */
    body: string | null

    /**
     * Parent commits of this commit.
     */
    parents: IGitCommit[]

    /**
     * The URL to this commit (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this commit (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this commit on its repository's external services.
     */
    externalURLs: IExternalLink[]

    /**
     * The Git tree in this commit at the given path.
     */
    tree: IGitTree | null

    /**
     * The Git blob in this commit at the given path.
     */
    blob: IGitBlob | null

    /**
     *  The file at the given path for this commit.
     *
     *  See "File" documentation for the difference between this field and the "blob" field.
     */
    file: File2 | null

    /**
     * Lists the programming languages present in the tree at this commit.
     */
    languages: string[]

    /**
     * The log of commits consisting of this commit and its ancestors.
     */
    ancestors: IGitCommitConnection

    /**
     * Returns the number of commits that this commit is behind and ahead of revspec.
     */
    behindAhead: IBehindAheadCounts

    /**
     * Symbols defined as of this commit. (All symbols, not just symbols that were newly defined in this commit.)
     */
    symbols: ISymbolConnection

    /**
     * Packages defined in this repository as of this commit, as returned by LSP workspace/xpackages
     * requests to this repository's language servers.
     */
    packages: IPackageConnection

    /**
     * Dependencies of this repository as of this commit, as returned by LSP workspace/xreferences
     * requests to this repository's language servers.
     */
    dependencies: IDependencyConnection
}

export interface ITreeOnGitCommitArguments {
    /**
     * The path of the tree.
     * @default
     */
    path?: string | null

    /**
     *  Whether to recurse into sub-trees. If true, it overrides the value of the "recursive" parameter on all of
     *  GitTree's fields.
     *
     *  DEPRECATED: Use the "recursive" parameter on GitTree's fields instead.
     * @default false
     */
    recursive?: boolean | null
}

export interface IBlobOnGitCommitArguments {
    path: string
}

export interface IFileOnGitCommitArguments {
    path: string
}

export interface IAncestorsOnGitCommitArguments {
    /**
     * Returns the first n commits from the list.
     */
    first?: number | null

    /**
     * Return commits that match the query.
     */
    query?: string | null

    /**
     * Return commits that affect the path.
     */
    path?: string | null
}

export interface IBehindAheadOnGitCommitArguments {
    revspec: string
}

export interface ISymbolsOnGitCommitArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

export interface IPackagesOnGitCommitArguments {
    /**
     * Returns the first n packages from the list.
     */
    first?: number | null

    /**
     * Return packages matching the query.
     */
    query?: string | null
}

export interface IDependenciesOnGitCommitArguments {
    /**
     * Returns the first n dependencies from the list.
     */
    first?: number | null

    /**
     * Return dependencies matching the query.
     */
    query?: string | null
}

/**
 * A signature.
 */
export interface ISignature {
    __typename: 'Signature'

    /**
     * The person.
     */
    person: IPerson

    /**
     * The date.
     */
    date: string
}

/**
 * A person.
 */
export interface IPerson {
    __typename: 'Person'

    /**
     * The name.
     */
    name: string

    /**
     * The email.
     */
    email: string

    /**
     * The name if set; otherwise the email username.
     */
    displayName: string

    /**
     * The avatar URL.
     */
    avatarURL: string

    /**
     * The corresponding user account for this person, if one exists.
     */
    user: IUser | null
}

/**
 * A user.
 */
export interface IUser {
    __typename: 'User'

    /**
     * The unique ID for the user.
     */
    id: ID

    /**
     * The user's username.
     */
    username: string

    /**
     * The unique numeric ID for the user.
     * @deprecated use id instead
     */
    sourcegraphID: number

    /**
     *  The user's primary email address.
     *
     *  Only the user and site admins can access this field.
     * @deprecated use emails instead
     */
    email: string

    /**
     * The display name chosen by the user.
     */
    displayName: string | null

    /**
     * The URL of the user's avatar image.
     */
    avatarURL: string | null

    /**
     * The URL to the user's profile on Sourcegraph.
     */
    url: string

    /**
     * The URL to the user's settings.
     */
    settingsURL: string

    /**
     * The date when the user account was created on Sourcegraph.
     */
    createdAt: string

    /**
     * The date when the user account was last updated on Sourcegraph.
     */
    updatedAt: string | null

    /**
     *  Whether the user is a site admin.
     *
     *  Only the user and site admins can access this field.
     */
    siteAdmin: boolean

    /**
     *  The latest settings for the user.
     *
     *  Only the user and site admins can access this field.
     */
    latestSettings: ISettings | null

    /**
     * The configuration cascade including this subject and all applicable subjects whose configuration is lower
     * precedence than this subject.
     */
    configurationCascade: IConfigurationCascade

    /**
     * The organizations that this user is a member of.
     */
    organizations: IOrgConnection

    /**
     * This user's organization memberships.
     */
    organizationMemberships: IOrganizationMembershipConnection

    /**
     *  Tags associated with the user. These are used for internal site management and feature selection.
     *
     *  Only the user and site admins can access this field.
     */
    tags: string[]

    /**
     *  The user's usage activity on Sourcegraph.
     *
     *  Only the user and site admins can access this field.
     */
    activity: IUserActivity

    /**
     *  The user's email addresses.
     *
     *  Only the user and site admins can access this field.
     */
    emails: IUserEmail[]

    /**
     *  The user's access tokens (which grant to the holder the privileges of the user). This consists
     *  of all access tokens whose subject is this user.
     *
     *  Only the user and site admins can access this field.
     */
    accessTokens: IAccessTokenConnection

    /**
     * A list of external accounts that are associated with the user.
     */
    externalAccounts: IExternalAccountConnection

    /**
     *  The user's currently active session.
     *
     *  Only the currently authenticated user can access this field. Site admins are not able to access sessions for
     *  other users.
     */
    session: ISession

    /**
     * Whether the viewer has admin privileges on this user. The user has admin privileges on their own user, and
     * site admins have admin privileges on all users.
     */
    viewerCanAdminister: boolean

    /**
     *  The user's survey responses.
     *
     *  Only the user and site admins can access this field.
     */
    surveyResponses: ISurveyResponse[]

    /**
     * A list of extensions published by this user in the extension registry.
     */
    registryExtensions: IRegistryExtensionConnection
}

export interface IAccessTokensOnUserArguments {
    /**
     * Returns the first n access tokens from the list.
     */
    first?: number | null
}

export interface IExternalAccountsOnUserArguments {
    /**
     * Returns the first n external accounts from the list.
     */
    first?: number | null
}

export interface IRegistryExtensionsOnUserArguments {
    /**
     * Returns the first n extensions from the list.
     */
    first?: number | null

    /**
     * Returns only extensions matching the query.
     */
    query?: string | null
}

/**
 * ConfigurationSubject is something that can have configuration.
 */
export type ConfigurationSubject = IUser | IOrg | ISite

/**
 * ConfigurationSubject is something that can have configuration.
 */
export interface IConfigurationSubject {
    __typename: 'ConfigurationSubject'

    /**
     * The ID.
     */
    id: ID

    /**
     * The latest settings.
     */
    latestSettings: ISettings | null

    /**
     * The URL to the settings.
     */
    settingsURL: string

    /**
     * Whether the viewer can modify the subject's configuration.
     */
    viewerCanAdminister: boolean

    /**
     * The configuration cascade including this subject and all applicable subjects whose configuration is lower
     * precedence than this subject.
     */
    configurationCascade: IConfigurationCascade
}

/**
 * Settings is a version of a configuration settings file.
 */
export interface ISettings {
    __typename: 'Settings'

    /**
     * The ID.
     */
    id: number

    /**
     * The configuration.
     */
    configuration: IConfiguration

    /**
     * The subject that these settings are for.
     */
    subject: ConfigurationSubject

    /**
     * The author.
     */
    author: IUser

    /**
     * The time when this was created.
     */
    createdAt: string

    /**
     * The contents.
     * @deprecated use configuration.contents instead
     */
    contents: string
}

/**
 * Configuration contains settings from (possibly) multiple settings files.
 */
export interface IConfiguration {
    __typename: 'Configuration'

    /**
     * The raw JSON contents, encoded as a string.
     */
    contents: string

    /**
     * Error and warning messages about the configuration.
     */
    messages: string[]
}

/**
 * The configurations for all of the relevant configuration subjects, plus the merged configuration.
 */
export interface IConfigurationCascade {
    __typename: 'ConfigurationCascade'

    /**
     * The default settings, which are applied first and the lowest priority behind
     * all configuration subjects' settings.
     */
    defaults: IConfiguration | null

    /**
     * The other configuration subjects that are applied with lower precedence than this subject to
     * form the final configuration. For example, a user in 2 organizations would have the following
     * configuration subjects: site (global settings), org 1, org 2, and the user.
     */
    subjects: ConfigurationSubject[]

    /**
     * The effective configuration, merged from all of the subjects.
     */
    merged: IConfiguration
}

/**
 * A list of organizations.
 */
export interface IOrgConnection {
    __typename: 'OrgConnection'

    /**
     * A list of organizations.
     */
    nodes: IOrg[]

    /**
     * The total count of organizations in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number
}

/**
 * An organization, which is a group of users.
 */
export interface IOrg {
    __typename: 'Org'

    /**
     * The unique ID for the organization.
     */
    id: ID

    /**
     * The organization's name. This is unique among all organizations on this Sourcegraph site.
     */
    name: string

    /**
     * The organization's chosen display name.
     */
    displayName: string | null

    /**
     * The date when the organization was created, in RFC 3339 format.
     */
    createdAt: string

    /**
     * A list of users who are members of this organization.
     */
    members: IUserConnection

    /**
     *  The latest settings for the organization.
     *
     *  Only organization members and site admins can access this field.
     */
    latestSettings: ISettings | null

    /**
     * The configuration cascade including this subject and all applicable subjects whose configuration is lower
     * precedence than this subject.
     */
    configurationCascade: IConfigurationCascade

    /**
     * A pending invitation for the viewer to join this organization, if any.
     */
    viewerPendingInvitation: IOrganizationInvitation | null

    /**
     * Whether the viewer has admin privileges on this organization. Currently, all of an organization's members
     * have admin privileges on the organization.
     */
    viewerCanAdminister: boolean

    /**
     * Whether the viewer is a member of this organization.
     */
    viewerIsMember: boolean

    /**
     * The URL to the organization.
     */
    url: string

    /**
     * The URL to the organization's settings.
     */
    settingsURL: string

    /**
     * A list of extensions published by this organization in the extension registry.
     */
    registryExtensions: IRegistryExtensionConnection
}

export interface IRegistryExtensionsOnOrgArguments {
    /**
     * Returns the first n extensions from the list.
     */
    first?: number | null

    /**
     * Returns only extensions matching the query.
     */
    query?: string | null
}

/**
 * A list of users.
 */
export interface IUserConnection {
    __typename: 'UserConnection'

    /**
     * A list of users.
     */
    nodes: IUser[]

    /**
     * The total count of users in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number
}

/**
 * An invitation to join an organization as a member.
 */
export interface IOrganizationInvitation {
    __typename: 'OrganizationInvitation'

    /**
     * The ID of the invitation.
     */
    id: ID

    /**
     * The organization that the invitation is for.
     */
    organization: IOrg

    /**
     * The user who sent the invitation.
     */
    sender: IUser

    /**
     * The user who received the invitation.
     */
    recipient: IUser

    /**
     * The date when this invitation was created.
     */
    createdAt: string

    /**
     * The most recent date when a notification was sent to the recipient about this invitation.
     */
    notifiedAt: string | null

    /**
     * The date when this invitation was responded to by the recipient.
     */
    respondedAt: string | null

    /**
     * The recipient's response to this invitation, or no response (null).
     */
    responseType: OrganizationInvitationResponseType | null

    /**
     * The URL where the recipient can respond to the invitation when pending, or null if not pending.
     */
    respondURL: string | null

    /**
     * The date when this invitation was revoked.
     */
    revokedAt: string | null
}

/**
 * The recipient's possible responses to an invitation to join an organization as a member.
 */
export enum OrganizationInvitationResponseType {
    /**
     * The invitation was accepted by the recipient.
     */
    ACCEPT = 'ACCEPT',

    /**
     * The invitation was rejected by the recipient.
     */
    REJECT = 'REJECT',
}

/**
 * A list of registry extensions.
 */
export interface IRegistryExtensionConnection {
    __typename: 'RegistryExtensionConnection'

    /**
     * A list of registry extensions.
     */
    nodes: IRegistryExtension[]

    /**
     * The total count of registry extensions in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * The URL to this list, or null if none exists.
     */
    url: string | null

    /**
     *  Errors that occurred while communicating with remote registries to obtain the list of extensions.
     *
     *  In order to be able to return local extensions even when the remote registry is unreachable, errors are
     *  recorded here instead of in the top-level GraphQL errors list.
     */
    error: string | null
}

/**
 * An extension's listing in the extension registry.
 */
export interface IRegistryExtension {
    __typename: 'RegistryExtension'

    /**
     * The unique, opaque, permanent ID of the extension. Do not display this ID to the user; display
     * RegistryExtension.extensionID instead (it is friendlier and still unique, but it can be renamed).
     */
    id: ID

    /**
     * The UUID of the extension. This identifies the extension externally (along with the origin). The UUID maps
     * 1-to-1 to RegistryExtension.id.
     */
    uuid: string

    /**
     * The publisher of the extension. If this extension is from a remote registry, the publisher may be null.
     */
    publisher: RegistryPublisher | null

    /**
     * The qualified, unique name that refers to this extension, consisting of the registry name (if non-default),
     * publisher's name, and the extension's name, all joined by "/" (for example, "acme-corp/my-extension-name").
     */
    extensionID: string

    /**
     * The extension ID without the registry name.
     */
    extensionIDWithoutRegistry: string

    /**
     * The name of the extension (not including the publisher's name).
     */
    name: string

    /**
     * The extension manifest, or null if none is set.
     */
    manifest: IExtensionManifest | null

    /**
     * The date when this extension was created on the registry.
     */
    createdAt: string | null

    /**
     * The date when this extension was last updated on the registry.
     */
    updatedAt: string | null

    /**
     * The URL to the extension on this Sourcegraph site.
     */
    url: string

    /**
     * The URL to the extension on the extension registry where it lives (if this is a remote
     * extension). If this extension is local, then this field's value is null.
     */
    remoteURL: string | null

    /**
     * The name of this extension's registry.
     */
    registryName: string

    /**
     * Whether the registry extension is published on this Sourcegraph site.
     */
    isLocal: boolean

    /**
     * Whether the viewer has admin privileges on this registry extension.
     */
    viewerCanAdminister: boolean
}

/**
 * A publisher of a registry extension.
 */
export type RegistryPublisher = IUser | IOrg

/**
 * A description of the extension, how to run or access it, and when to activate it.
 */
export interface IExtensionManifest {
    __typename: 'ExtensionManifest'

    /**
     * The raw JSON contents of the manifest.
     */
    raw: string

    /**
     * The title specified in the manifest, if any.
     */
    title: string | null

    /**
     * The description specified in the manifest, if any.
     */
    description: string | null
}

/**
 * Pagination information. See https://facebook.github.io/relay/graphql/connections.htm#sec-undefined.PageInfo.
 */
export interface IPageInfo {
    __typename: 'PageInfo'

    /**
     * Whether there is a next page of nodes in the connection.
     */
    hasNextPage: boolean
}

/**
 * A list of organization memberships.
 */
export interface IOrganizationMembershipConnection {
    __typename: 'OrganizationMembershipConnection'

    /**
     * A list of organization memberships.
     */
    nodes: IOrganizationMembership[]

    /**
     * The total count of organization memberships in the connection. This total count may be larger than the number
     * of nodes in this object when the result is paginated.
     */
    totalCount: number
}

/**
 * An organization membership.
 */
export interface IOrganizationMembership {
    __typename: 'OrganizationMembership'

    /**
     * The organization.
     */
    organization: IOrg

    /**
     * The user.
     */
    user: IUser

    /**
     * The time when this was created.
     */
    createdAt: string

    /**
     * The time when this was updated.
     */
    updatedAt: string
}

/**
 * UserActivity describes a user's activity on the site.
 */
export interface IUserActivity {
    __typename: 'UserActivity'

    /**
     * The number of search queries that the user has performed.
     */
    searchQueries: number

    /**
     * The number of page views that the user has performed.
     */
    pageViews: number

    /**
     * The number of code intelligence actions that the user has performed.
     */
    codeIntelligenceActions: number

    /**
     * The last time the user was active (any action, any platform).
     */
    lastActiveTime: string | null

    /**
     * The last time the user was active on a code host integration.
     */
    lastActiveCodeHostIntegrationTime: string | null
}

/**
 * A user's email address.
 */
export interface IUserEmail {
    __typename: 'UserEmail'

    /**
     * The email address.
     */
    email: string

    /**
     * Whether the email address has been verified by the user.
     */
    verified: boolean

    /**
     * Whether the email address is pending verification.
     */
    verificationPending: boolean

    /**
     * The user associated with this email address.
     */
    user: IUser

    /**
     * Whether the viewer has privileges to manually mark this email address as verified (without the user going
     * through the normal verification process). Only site admins have this privilege.
     */
    viewerCanManuallyVerify: boolean
}

/**
 * A list of access tokens.
 */
export interface IAccessTokenConnection {
    __typename: 'AccessTokenConnection'

    /**
     * A list of access tokens.
     */
    nodes: IAccessToken[]

    /**
     * The total count of access tokens in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * An access token that grants to the holder the privileges of the user who created it.
 */
export interface IAccessToken {
    __typename: 'AccessToken'

    /**
     * The unique ID for the access token.
     */
    id: ID

    /**
     * The user whose privileges the access token grants.
     */
    subject: IUser

    /**
     * The scopes that define the allowed set of operations that can be performed using this access token.
     */
    scopes: string[]

    /**
     * A user-supplied descriptive note for the access token.
     */
    note: string

    /**
     * The user who created the access token. This is either the subject user (if the access token
     * was created by the same user) or a site admin (who can create access tokens for any user).
     */
    creator: IUser

    /**
     * The date when the access token was created.
     */
    createdAt: string

    /**
     * The date when the access token was last used to authenticate a request.
     */
    lastUsedAt: string | null
}

/**
 * A list of external accounts.
 */
export interface IExternalAccountConnection {
    __typename: 'ExternalAccountConnection'

    /**
     * A list of external accounts.
     */
    nodes: IExternalAccount[]

    /**
     * The total count of external accounts in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * An external account associated with a user.
 */
export interface IExternalAccount {
    __typename: 'ExternalAccount'

    /**
     * The unique ID for the external account.
     */
    id: ID

    /**
     * The user on Sourcegraph.
     */
    user: IUser

    /**
     * The type of the external service where the external account resides.
     */
    serviceType: string

    /**
     * An identifier for the external service where the external account resides.
     */
    serviceID: string

    /**
     * An identifier for the client of the external service where the external account resides. This distinguishes
     * among multiple authentication providers that access the same service with different parameters.
     */
    clientID: string

    /**
     * An identifier for the external account (typically equal to or derived from the ID on the external service).
     */
    accountID: string

    /**
     * The creation date of this external account on Sourcegraph.
     */
    createdAt: string

    /**
     * The last-updated date of this external account on Sourcegraph.
     */
    updatedAt: string

    /**
     * A URL that, when visited, re-initiates the authentication process.
     */
    refreshURL: string | null

    /**
     *  Provider-specific data about the external account.
     *
     *  Only site admins may query this field.
     */
    accountData: any | null
}

/**
 * An active user session.
 */
export interface ISession {
    __typename: 'Session'

    /**
     * Whether the user can sign out of this session on Sourcegraph.
     */
    canSignOut: boolean
}

/**
 * An individual response to a user satisfaction (NPS) survey.
 */
export interface ISurveyResponse {
    __typename: 'SurveyResponse'

    /**
     * The unique ID of the survey response
     */
    id: ID

    /**
     * The user who submitted the survey (if they were authenticated at the time).
     */
    user: IUser | null

    /**
     * The email that the user manually entered (if they were NOT authenticated at the time).
     */
    email: string | null

    /**
     * User's likelihood of recommending Sourcegraph to a friend, from 0-10.
     */
    score: number

    /**
     * The answer to "What is the most important reason for the score you gave".
     */
    reason: string | null

    /**
     * The answer to "What can Sourcegraph do to provide a better product"
     */
    better: string | null

    /**
     * The time when this response was created.
     */
    createdAt: string
}

/**
 * A URL to a resource on an external service, such as the URL to a repository on its external (origin) code host.
 */
export interface IExternalLink {
    __typename: 'ExternalLink'

    /**
     * The URL to the resource.
     */
    url: string

    /**
     * The type of external service, such as "github", or null if unknown/unrecognized. This is used solely for
     * displaying an icon that represents the service.
     */
    serviceType: string | null
}

/**
 * A Git tree in a repository.
 */
export interface IGitTree {
    __typename: 'GitTree'

    /**
     * The full path (relative to the root) of this tree.
     */
    path: string

    /**
     * Whether this tree is the root (top-level) tree.
     */
    isRoot: boolean

    /**
     * The base name (i.e., last path component only) of this tree.
     */
    name: string

    /**
     * True because this is a directory. (The value differs for other TreeEntry interface implementations, such as
     * File.)
     */
    isDirectory: boolean

    /**
     * The Git commit containing this tree.
     */
    commit: IGitCommit

    /**
     * The repository containing this tree.
     */
    repository: IRepository

    /**
     * The URL to this tree (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this tree (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this tree on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * A list of directories in this tree.
     */
    directories: IGitTree[]

    /**
     * A list of files in this tree.
     */
    files: IFile[]

    /**
     * A list of entries in this tree.
     */
    entries: TreeEntry[]

    /**
     * Symbols defined in this tree.
     */
    symbols: ISymbolConnection
}

export interface IDirectoriesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null
}

export interface IFilesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null
}

export interface IEntriesOnGitTreeArguments {
    /**
     * Returns the first n files in the tree.
     */
    first?: number | null

    /**
     * Recurse into sub-trees.
     * @default false
     */
    recursive?: boolean | null
}

export interface ISymbolsOnGitTreeArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

/**
 * A file, directory, or other tree entry.
 */
export type TreeEntry = IGitTree | IGitBlob

/**
 * A file, directory, or other tree entry.
 */
export interface ITreeEntry {
    __typename: 'TreeEntry'

    /**
     * The full path (relative to the repository root) of this tree entry.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this tree entry.
     */
    name: string

    /**
     * Whether this tree entry is a directory.
     */
    isDirectory: boolean

    /**
     * The URL to this tree entry (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this tree entry (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this tree entry on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Symbols defined in this file or directory.
     */
    symbols: ISymbolConnection
}

export interface ISymbolsOnTreeEntryArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

/**
 * A list of symbols.
 */
export interface ISymbolConnection {
    __typename: 'SymbolConnection'

    /**
     * A list of symbols.
     */
    nodes: ISymbol[]

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 *  A code symbol (e.g., a function, variable, type, class, etc.).
 *
 *  It is derived from symbols as defined in the Language Server Protocol (see
 *  https://microsoft.github.io/language-server-protocol/specification#workspace_symbol).
 */
export interface ISymbol {
    __typename: 'Symbol'

    /**
     * The name of the symbol.
     */
    name: string

    /**
     * The name of the symbol that contains this symbol, if any. This field's value is not guaranteed to be
     * structured in such a way that callers can infer a hierarchy of symbols.
     */
    containerName: string | null

    /**
     * The kind of the symbol.
     */
    kind: SymbolKind

    /**
     * The programming language of the symbol.
     */
    language: string

    /**
     * The location where this symbol is defined.
     */
    location: ILocation

    /**
     * The URL to this symbol (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this symbol (using an immutable revision specifier).
     */
    canonicalURL: string
}

/**
 * All possible kinds of symbols. This set matches that of the Language Server Protocol
 * (https://microsoft.github.io/language-server-protocol/specification#workspace_symbol).
 */
export enum SymbolKind {
    UNKNOWN = 'UNKNOWN',
    FILE = 'FILE',
    MODULE = 'MODULE',
    NAMESPACE = 'NAMESPACE',
    PACKAGE = 'PACKAGE',
    CLASS = 'CLASS',
    METHOD = 'METHOD',
    PROPERTY = 'PROPERTY',
    FIELD = 'FIELD',
    CONSTRUCTOR = 'CONSTRUCTOR',
    ENUM = 'ENUM',
    INTERFACE = 'INTERFACE',
    FUNCTION = 'FUNCTION',
    VARIABLE = 'VARIABLE',
    CONSTANT = 'CONSTANT',
    STRING = 'STRING',
    NUMBER = 'NUMBER',
    BOOLEAN = 'BOOLEAN',
    ARRAY = 'ARRAY',
    OBJECT = 'OBJECT',
    KEY = 'KEY',
    NULL = 'NULL',
    ENUMMEMBER = 'ENUMMEMBER',
    STRUCT = 'STRUCT',
    EVENT = 'EVENT',
    OPERATOR = 'OPERATOR',
    TYPEPARAMETER = 'TYPEPARAMETER',
}

/**
 * A location inside a resource (in a repository at a specific commit).
 */
export interface ILocation {
    __typename: 'Location'

    /**
     * The file that this location refers to.
     */
    resource: IGitBlob

    /**
     * The range inside the file that this location refers to.
     */
    range: IRange | null

    /**
     * The URL to this location (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this location (using an immutable revision specifier).
     */
    canonicalURL: string
}

/**
 * A Git blob in a repository.
 */
export interface IGitBlob {
    __typename: 'GitBlob'

    /**
     * The full path (relative to the repository root) of this blob.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this blob's path.
     */
    name: string

    /**
     * False because this is a blob (file), not a directory.
     */
    isDirectory: boolean

    /**
     * The content of this blob.
     */
    content: string

    /**
     * Whether or not it is binary.
     */
    binary: boolean

    /**
     *  The blob contents rendered as rich HTML, or an empty string if it is not a supported
     *  rich file type.
     *
     *  This HTML string is already escaped and thus is always safe to render.
     */
    richHTML: string

    /**
     * The Git commit containing this blob.
     */
    commit: IGitCommit

    /**
     * The repository containing this Git blob.
     */
    repository: IRepository

    /**
     * The URL to this blob (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this blob (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this blob on its repository's external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Blame the blob.
     */
    blame: IHunk[]

    /**
     * Highlight the blob contents.
     */
    highlight: IHighlightedFile

    /**
     * Returns dependency references for the blob.
     */
    dependencyReferences: IDependencyReferences

    /**
     * Symbols defined in this blob.
     */
    symbols: ISymbolConnection
}

export interface IBlameOnGitBlobArguments {
    startLine: number
    endLine: number
}

export interface IHighlightOnGitBlobArguments {
    disableTimeout: boolean
    isLightTheme: boolean
}

export interface IDependencyReferencesOnGitBlobArguments {
    Language: string
    Line: number
    Character: number
}

export interface ISymbolsOnGitBlobArguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

/**
 *  A file.
 *
 *  In a future version of Sourcegraph, a repository's files may be distinct from a repository's blobs
 *  (for example, to support searching/browsing generated files that aren't committed and don't exist
 *  as Git blobs). Clients should generally use the GitBlob concrete type and GitCommit.blobs (not
 *  GitCommit.files), unless they explicitly want to opt-in to different behavior in the future.
 *
 *  INTERNAL: This is temporarily named File2 during a migration. Do not refer to the name File2 in
 *  any API clients as the name will change soon.
 */
export type File2 = IGitBlob

/**
 *  A file.
 *
 *  In a future version of Sourcegraph, a repository's files may be distinct from a repository's blobs
 *  (for example, to support searching/browsing generated files that aren't committed and don't exist
 *  as Git blobs). Clients should generally use the GitBlob concrete type and GitCommit.blobs (not
 *  GitCommit.files), unless they explicitly want to opt-in to different behavior in the future.
 *
 *  INTERNAL: This is temporarily named File2 during a migration. Do not refer to the name File2 in
 *  any API clients as the name will change soon.
 */
export interface IFile2 {
    __typename: 'File2'

    /**
     * The full path (relative to the root) of this file.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this file.
     */
    name: string

    /**
     * False because this is a file, not a directory.
     */
    isDirectory: boolean

    /**
     * The content of this file.
     */
    content: string

    /**
     * Whether or not it is binary.
     */
    binary: boolean

    /**
     *  The file rendered as rich HTML, or an empty string if it is not a supported
     *  rich file type.
     *
     *  This HTML string is already escaped and thus is always safe to render.
     */
    richHTML: string

    /**
     * The URL to this file (using the input revision specifier, which may not be immutable).
     */
    url: string

    /**
     * The canonical URL to this file (using an immutable revision specifier).
     */
    canonicalURL: string

    /**
     * The URLs to this file on external services.
     */
    externalURLs: IExternalLink[]

    /**
     * Highlight the file.
     */
    highlight: IHighlightedFile

    /**
     * Returns dependency references for the file.
     */
    dependencyReferences: IDependencyReferences

    /**
     * Symbols defined in this file.
     */
    symbols: ISymbolConnection
}

export interface IHighlightOnFile2Arguments {
    disableTimeout: boolean
    isLightTheme: boolean
}

export interface IDependencyReferencesOnFile2Arguments {
    Language: string
    Line: number
    Character: number
}

export interface ISymbolsOnFile2Arguments {
    /**
     * Returns the first n symbols from the list.
     */
    first?: number | null

    /**
     * Return symbols matching the query.
     */
    query?: string | null
}

/**
 * A highlighted file.
 */
export interface IHighlightedFile {
    __typename: 'HighlightedFile'

    /**
     * Whether or not it was aborted.
     */
    aborted: boolean

    /**
     * The HTML.
     */
    html: string
}

/**
 * Dependency references.
 */
export interface IDependencyReferences {
    __typename: 'DependencyReferences'

    /**
     * The dependency reference data.
     */
    dependencyReferenceData: IDependencyReferencesData

    /**
     * The repository data.
     */
    repoData: IRepoDataMap
}

/**
 * Dependency references data.
 */
export interface IDependencyReferencesData {
    __typename: 'DependencyReferencesData'

    /**
     * The references.
     */
    references: IDependencyReference[]

    /**
     * The location.
     */
    location: IDepLocation
}

/**
 * A dependency reference.
 */
export interface IDependencyReference {
    __typename: 'DependencyReference'

    /**
     * The dependency data.
     */
    dependencyData: string

    /**
     * The repository ID.
     */
    repoId: number

    /**
     * The hints.
     */
    hints: string
}

/**
 * A dependency location.
 */
export interface IDepLocation {
    __typename: 'DepLocation'

    /**
     * The location.
     */
    location: string

    /**
     * The symbol.
     */
    symbol: string
}

/**
 * A repository data map.
 */
export interface IRepoDataMap {
    __typename: 'RepoDataMap'

    /**
     * The repositories.
     */
    repos: IRepository[]

    /**
     * The repository IDs.
     */
    repoIds: number[]
}

/**
 * A hunk.
 */
export interface IHunk {
    __typename: 'Hunk'

    /**
     * The startLine.
     */
    startLine: number

    /**
     * The endLine.
     */
    endLine: number

    /**
     * The startByte.
     */
    startByte: number

    /**
     * The endByte.
     */
    endByte: number

    /**
     * The rev.
     */
    rev: string

    /**
     * The author.
     */
    author: ISignature

    /**
     * The message.
     */
    message: string

    /**
     * The commit that contains the hunk.
     */
    commit: IGitCommit
}

/**
 * A range inside a file. The start position is inclusive, and the end position is exclusive.
 */
export interface IRange {
    __typename: 'Range'

    /**
     * The start position of the range (inclusive).
     */
    start: IPosition

    /**
     * The end position of the range (exclusive).
     */
    end: IPosition
}

/**
 * A zero-based position inside a file.
 */
export interface IPosition {
    __typename: 'Position'

    /**
     * The line number (zero-based) of the position.
     */
    line: number

    /**
     * The character offset (zero-based) in the line of the position.
     */
    character: number
}

/**
 * File is temporarily preserved for backcompat with browser extension search API client code.
 */
export interface IFile {
    __typename: 'File'

    /**
     * The full path (relative to the repository root) of this file.
     */
    path: string

    /**
     * The base name (i.e., file name only) of this file's path.
     */
    name: string

    /**
     * Whether this is a directory.
     */
    isDirectory: boolean

    /**
     * The URL to this file on Sourcegraph.
     */
    url: string

    /**
     * The repository that contains this file.
     */
    repository: IRepository
}

/**
 * A list of Git commits.
 */
export interface IGitCommitConnection {
    __typename: 'GitCommitConnection'

    /**
     * A list of Git commits.
     */
    nodes: IGitCommit[]

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A set of Git behind/ahead counts for one commit relative to another.
 */
export interface IBehindAheadCounts {
    __typename: 'BehindAheadCounts'

    /**
     * The number of commits behind the other commit.
     */
    behind: number

    /**
     * The number of commits ahead of the other commit.
     */
    ahead: number
}

/**
 * A list of packages.
 */
export interface IPackageConnection {
    __typename: 'PackageConnection'

    /**
     * A list of packages.
     */
    nodes: IPackage[]

    /**
     * The total count of packages in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 *  A package represents a grouping of code that is returned by a language server in response to a
 *  workspace/xpackages request.
 *
 *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
 */
export interface IPackage {
    __typename: 'Package'

    /**
     * The ID of the package.
     */
    id: ID

    /**
     * The repository commit that defines the package.
     */
    definingCommit: IGitCommit

    /**
     * The programming language used to define the package.
     */
    language: string

    /**
     *  The package descriptor, as returned by the language server's workspace/xpackages LSP method. The attribute
     *  names and values are defined by each language server and should generally be considered opaque.
     *
     *  The ordering is not meaningful.
     *
     *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
     */
    data: IKeyValue[]

    /**
     *  This package's dependencies, as returned by the language server's workspace/xpackages LSP method.
     *
     *  The ordering is not meaningful.
     *
     *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
     */
    dependencies: IDependency[]

    /**
     *  The list of references (from only this repository at the definingCommit) to definitions in this package.
     *
     *  If this operation is not supported (by the language server), this field's value will be null.
     */
    internalReferences: IReferenceConnection | null

    /**
     *  The list of references (from other repositories' packages) to definitions in this package. Currently this
     *  lists packages that refer to this package, NOT individual call/reference sites within those referencing
     *  packages (unlike internalReferences, which does list individual call sites). If this operation is not
     *  supported (by the language server), this field's value will be null.
     *
     *  EXPERIMENTAL: This field is experimental. It is subject to change. Please report any issues you see, and
     *  contact support for help.
     */
    externalReferences: IReferenceConnection | null
}

/**
 * A key-value pair.
 */
export interface IKeyValue {
    __typename: 'KeyValue'

    /**
     * The key.
     */
    key: string

    /**
     * The value, which can be of any type.
     */
    value: any
}

/**
 *  A dependency represents a dependency relationship between two units of code. It is derived from data returned by
 *  a language server in response to a workspace/xreferences request.
 *
 *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
 */
export interface IDependency {
    __typename: 'Dependency'

    /**
     * The ID of the dependency.
     */
    id: ID

    /**
     * The repository commit that depends on the unit of code described by this resolver's other fields.
     */
    dependingCommit: IGitCommit

    /**
     * The programming language of the dependency.
     */
    language: string

    /**
     *  The dependency attributes, as returned by the language server's workspace/xdependencies LSP method. The
     *  attribute names and values are defined by each language server and should generally be considered opaque.
     *  They generally overlap with the package descriptor's fields in the Package type.
     *
     *  The ordering is not meaningful.
     *
     *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
     */
    data: IKeyValue[]

    /**
     *  Hints that can be passed to workspace/xreferences to speed up retrieval of references to this dependency.
     *  These hints are returned by the language server's workspace/xdependencies LSP method. The attribute names and
     *  values are defined by each language server and should generally be considered opaque.
     *
     *  The ordering is not meaningful.
     *
     *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
     */
    hints: IKeyValue[]

    /**
     *  The list of references (in the depending commit's code files) to definitions in this dependency.
     *
     *  If this operation is not supported (by the language server), this field's value will be null.
     *
     *  EXPERIMENTAL: This field is experimental. It is subject to change. Please report any issues you see, and
     *  contact support for help.
     */
    references: IReferenceConnection | null
}

/**
 *  A list of code references (e.g., function calls, variable references, package import statements, etc.), as
 *  returned by language server(s) over LSP.
 *
 *  NOTE: The actual references (which would be expected to be available in the "nodes" field) are not exposed. This
 *  is because currently there are no API consumers that need them. In the future, they will be available here, but
 *  in the meantime, consumers can provide the searchQuery to the Query.search GraphQL resolver to retrieve
 *  references.
 */
export interface IReferenceConnection {
    __typename: 'ReferenceConnection'

    /**
     * The total count of references in this connection. If an exact count is not available, then this field's value
     * will be null; consult the approximateCount field instead.
     */
    totalCount: number | null

    /**
     * The approximate count of references in this connection. If counting is not supported, then this field's value
     * will be null.
     */
    approximateCount: IApproximateCount | null

    /**
     *  The search query (for Sourcegraph search) that matches references in this connection.
     *
     *  The query string does not include any repo:REPO@REV tokens (even if this connection would seem to warrant
     *  the inclusion of such tokens). Therefore, clients must add those tokens if they wish to constrain the search
     *  to only certain repositories and revisions. (This is so that clients can use the nice revision instead of the
     *  40-character Git commit SHA if desired.)
     */
    queryString: string

    /**
     *  The symbol descriptor query to pass to language servers in the LSP workspace/xreferences request to retrieve
     *  all references in this connection. This is derived from the attributes data of this connection's subject
     *  (e.g., Package.data or Dependency.data). The attribute names and values are defined by each language server
     *  and should generally be considered opaque.
     *
     *  The ordering is not meaningful.
     *
     *  See https://github.com/sourcegraph/language-server-protocol/blob/master/extension-workspace-references.md.
     */
    symbolDescriptor: IKeyValue[]
}

/**
 * An approximate count. To display this to the user, use ApproximateCount.label as the number and use
 * ApproximateCount.count to determine whether to pluralize the noun (if any) adjacent to the label.
 */
export interface IApproximateCount {
    __typename: 'ApproximateCount'

    /**
     * The count, which may be inexact. This number is always the prefix of the label field.
     */
    count: number

    /**
     * Whether the count finished and is exact.
     */
    exact: boolean

    /**
     * A textual label that approximates the count (e.g., "99+" if the counting is cut off at 99).
     */
    label: string
}

/**
 * A list of dependencies.
 */
export interface IDependencyConnection {
    __typename: 'DependencyConnection'

    /**
     * A list of dependencies.
     */
    nodes: IDependency[]

    /**
     * The total count of dependencies in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * Information and status about the mirroring of a repository. In this case, the remote source repository
 * is external to Sourcegraph and the mirror is maintained by the Sourcegraph site (not the other way
 * around).
 */
export interface IMirrorRepositoryInfo {
    __typename: 'MirrorRepositoryInfo'

    /**
     * The URL of the remote source repository.
     */
    remoteURL: string

    /**
     * Whether the clone of the repository has begun but not yet completed.
     */
    cloneInProgress: boolean

    /**
     * A single line of text that contains progress information for the running clone command.
     * The format of the progress text is not specified.
     * It is intended to be displayed directly to a user.
     * e.g.
     * "Receiving objects:  95% (2041/2148), 292.01 KiB | 515.00 KiB/s"
     * "Resolving deltas:   9% (117/1263)"
     */
    cloneProgress: string | null

    /**
     * Whether the repository has ever been successfully cloned.
     */
    cloned: boolean

    /**
     * When the repository was last successfully updated from the remote source repository..
     */
    updatedAt: string | null
}

/**
 * A repository on an external service (such as GitHub, GitLab, Phabricator, etc.).
 */
export interface IExternalRepository {
    __typename: 'ExternalRepository'

    /**
     *  The repository's ID on the external service.
     *
     *  Example: For GitHub, this is the GitHub GraphQL API's node ID for the repository.
     */
    id: string

    /**
     *  The type of external service where this repository resides.
     *
     *  Example: "github", "gitlab", etc.
     */
    serviceType: string

    /**
     *  The particular instance of the external service where this repository resides. Its value is
     *  opaque but typically consists of the canonical base URL to the service.
     *
     *  Example: For GitHub.com, this is "https://github.com/".
     */
    serviceID: string
}

/**
 * Information about a repository's text search index.
 */
export interface IRepositoryTextSearchIndex {
    __typename: 'RepositoryTextSearchIndex'

    /**
     * The indexed repository.
     */
    repository: IRepository

    /**
     * The status of the text search index, if available.
     */
    status: IRepositoryTextSearchIndexStatus | null

    /**
     * Git refs in the repository that are configured for text search indexing.
     */
    refs: IRepositoryTextSearchIndexedRef[]
}

/**
 * The status of a repository's text search index.
 */
export interface IRepositoryTextSearchIndexStatus {
    __typename: 'RepositoryTextSearchIndexStatus'

    /**
     * The date that the index was last updated.
     */
    updatedAt: string

    /**
     * The byte size of the original content.
     */
    contentByteSize: number

    /**
     * The number of files in the original content.
     */
    contentFilesCount: number

    /**
     * The byte size of the index.
     */
    indexByteSize: number

    /**
     * The number of index shards.
     */
    indexShardsCount: number
}

/**
 * A Git ref (usually a branch) in a repository that is configured to be indexed for text search.
 */
export interface IRepositoryTextSearchIndexedRef {
    __typename: 'RepositoryTextSearchIndexedRef'

    /**
     * The Git ref (usually a branch) that is configured to be indexed for text search. To find the specific commit
     * SHA that was indexed, use RepositoryTextSearchIndexedRef.indexedCommit; this field's ref target resolves to
     * the current target, not the target at the time of indexing.
     */
    ref: IGitRef

    /**
     * Whether a text search index exists for this ref.
     */
    indexed: boolean

    /**
     * Whether the text search index is of the current commit for the Git ref. If false, the index is stale.
     */
    current: boolean

    /**
     * The indexed Git commit (which may differ from the ref's current target if the index is out of date). If
     * indexed is false, this field's value is null.
     */
    indexedCommit: IGitObject | null
}

/**
 * A Git ref.
 */
export interface IGitRef {
    __typename: 'GitRef'

    /**
     * The globally addressable ID for the Git ref.
     */
    id: ID

    /**
     * The full ref name (e.g., "refs/heads/mybranch" or "refs/tags/mytag").
     */
    name: string

    /**
     * An unambiguous short name for the ref.
     */
    abbrevName: string

    /**
     *  The display name of the ref. For branches ("refs/heads/foo"), this is the branch
     *  name ("foo").
     *
     *  As a special case, for GitHub pull request refs of the form refs/pull/NUMBER/head,
     *  this is "#NUMBER".
     */
    displayName: string

    /**
     * The prefix of the ref, either "", "refs/", "refs/heads/", "refs/pull/", or
     * "refs/tags/". This prefix is always a prefix of the ref's name.
     */
    prefix: string

    /**
     * The type of this Git ref.
     */
    type: GitRefType

    /**
     * The object that the ref points to.
     */
    target: IGitObject

    /**
     * The associated repository.
     */
    repository: IRepository

    /**
     * The URL to this Git ref.
     */
    url: string
}

/**
 * All possible types of Git refs.
 */
export enum GitRefType {
    /**
     * A Git branch (in refs/heads/).
     */
    GIT_BRANCH = 'GIT_BRANCH',

    /**
     * A Git tag (in refs/tags/).
     */
    GIT_TAG = 'GIT_TAG',

    /**
     * A Git ref that is neither a branch nor tag.
     */
    GIT_REF_OTHER = 'GIT_REF_OTHER',
}

/**
 * A Git object.
 */
export interface IGitObject {
    __typename: 'GitObject'

    /**
     * This object's OID.
     */
    oid: any

    /**
     * The abbreviated form of this object's OID.
     */
    abbreviatedOID: string

    /**
     * The commit object, if it is a commit and it exists; otherwise null.
     */
    commit: IGitCommit | null

    /**
     * The Git object's type.
     */
    type: GitObjectType
}

/**
 * All possible types of Git objects.
 */
export enum GitObjectType {
    /**
     * A Git commit object.
     */
    GIT_COMMIT = 'GIT_COMMIT',

    /**
     * A Git tag object.
     */
    GIT_TAG = 'GIT_TAG',

    /**
     * A Git tree object.
     */
    GIT_TREE = 'GIT_TREE',

    /**
     * A Git blob object.
     */
    GIT_BLOB = 'GIT_BLOB',

    /**
     * A Git object of unknown type.
     */
    GIT_UNKNOWN = 'GIT_UNKNOWN',
}

/**
 * Ordering options for Git refs.
 */
export enum GitRefOrder {
    /**
     * By the authored or committed at date, whichever is more recent.
     */
    AUTHORED_OR_COMMITTED_AT = 'AUTHORED_OR_COMMITTED_AT',
}

/**
 * A list of Git refs.
 */
export interface IGitRefConnection {
    __typename: 'GitRefConnection'

    /**
     * A list of Git refs.
     */
    nodes: IGitRef[]

    /**
     * The total count of Git refs in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * The differences between two Git commits in a repository.
 */
export interface IRepositoryComparison {
    __typename: 'RepositoryComparison'

    /**
     * The range that this comparison represents.
     */
    range: IGitRevisionRange

    /**
     * The commits in the comparison range, excluding the base and including the head.
     */
    commits: IGitCommitConnection

    /**
     * The file diffs for each changed file.
     */
    fileDiffs: IFileDiffConnection
}

export interface ICommitsOnRepositoryComparisonArguments {
    /**
     * Return the first n commits from the list.
     */
    first?: number | null
}

export interface IFileDiffsOnRepositoryComparisonArguments {
    /**
     * Return the first n file diffs from the list.
     */
    first?: number | null
}

/**
 * A Git revision range of the form "base..head" or "base...head". Other revision
 * range formats are not supported.
 */
export interface IGitRevisionRange {
    __typename: 'GitRevisionRange'

    /**
     * The Git revision range expression of the form "base..head" or "base...head".
     */
    expr: string

    /**
     * The base (left-hand side) of the range.
     */
    base: GitRevSpec

    /**
     * The base's revspec as an expression.
     */
    baseRevSpec: IGitRevSpecExpr

    /**
     * The head (right-hand side) of the range.
     */
    head: GitRevSpec

    /**
     * The head's revspec as an expression.
     */
    headRevSpec: IGitRevSpecExpr

    /**
     * The merge-base of the base and head revisions, if this is a "base...head"
     * revision range. If this is a "base..head" revision range, then this field is null.
     */
    mergeBase: IGitObject | null
}

/**
 * A Git revspec.
 */
export type GitRevSpec = IGitRef | IGitRevSpecExpr | IGitObject

/**
 * A Git revspec expression that (possibly) resolves to a Git revision.
 */
export interface IGitRevSpecExpr {
    __typename: 'GitRevSpecExpr'

    /**
     * The original Git revspec expression.
     */
    expr: string

    /**
     * The Git object that the revspec resolves to, or null otherwise.
     */
    object: IGitObject | null
}

/**
 * A list of file diffs.
 */
export interface IFileDiffConnection {
    __typename: 'FileDiffConnection'

    /**
     * A list of file diffs.
     */
    nodes: IFileDiff[]

    /**
     * The total count of file diffs in the connection, if available. This total count may be larger than the number
     * of nodes in this object when the result is paginated.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo

    /**
     * The diff stat for the file diffs in this object, which may be a subset of the entire diff if the result is
     * paginated.
     */
    diffStat: IDiffStat

    /**
     * The raw diff for the file diffs in this object, which may be a subset of the entire diff if the result is
     * paginated.
     */
    rawDiff: string
}

/**
 * A diff for a single file.
 */
export interface IFileDiff {
    __typename: 'FileDiff'

    /**
     * The old (original) path of the file, or null if the file was added.
     */
    oldPath: string | null

    /**
     * The old file, or null if the file was created (oldFile.path == oldPath).
     */
    oldFile: File2 | null

    /**
     * The new (changed) path of the file, or null if the file was deleted.
     */
    newPath: string | null

    /**
     * The new file, or null if the file was deleted (newFile.path == newPath).
     */
    newFile: File2 | null

    /**
     * The old file (if the file was deleted) and otherwise the new file. This file field is typically used by
     * clients that want to show a "View" link to the file.
     */
    mostRelevantFile: File2

    /**
     * Hunks that were changed from old to new.
     */
    hunks: IFileDiffHunk[]

    /**
     * The diff stat for the whole file.
     */
    stat: IDiffStat

    /**
     *  FOR INTERNAL USE ONLY.
     *
     *  An identifier for the file diff that is unique among all other file diffs in the list that
     *  contains it.
     */
    internalID: string
}

/**
 * A changed region ("hunk") in a file diff.
 */
export interface IFileDiffHunk {
    __typename: 'FileDiffHunk'

    /**
     * The range of the old file that the hunk applies to.
     */
    oldRange: IFileDiffHunkRange

    /**
     * Whether the old file had a trailing newline.
     */
    oldNoNewlineAt: boolean

    /**
     * The range of the new file that the hunk applies to.
     */
    newRange: IFileDiffHunkRange

    /**
     * The diff hunk section heading, if any.
     */
    section: string | null

    /**
     * The hunk body, with lines prefixed with '-', '+', or ' '.
     */
    body: string
}

/**
 * A hunk range in one side (old/new) of a diff.
 */
export interface IFileDiffHunkRange {
    __typename: 'FileDiffHunkRange'

    /**
     * The first line that the hunk applies to.
     */
    startLine: number

    /**
     * The number of lines that the hunk applies to.
     */
    lines: number
}

/**
 * Statistics about a diff.
 */
export interface IDiffStat {
    __typename: 'DiffStat'

    /**
     * Number of additions.
     */
    added: number

    /**
     * Number of changes.
     */
    changed: number

    /**
     * Number of deletions.
     */
    deleted: number
}

/**
 * A list of contributors to a repository.
 */
export interface IRepositoryContributorConnection {
    __typename: 'RepositoryContributorConnection'

    /**
     * A list of contributors to a repository.
     */
    nodes: IRepositoryContributor[]

    /**
     * The total count of contributors in the connection, if available. This total count may be larger than the
     * number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A contributor to a repository.
 */
export interface IRepositoryContributor {
    __typename: 'RepositoryContributor'

    /**
     * The personal information for the contributor.
     */
    person: IPerson

    /**
     * The number of contributions made by this contributor.
     */
    count: number

    /**
     * The repository in which the contributions occurred.
     */
    repository: IRepository

    /**
     * Commits by the contributor.
     */
    commits: IGitCommitConnection
}

export interface ICommitsOnRepositoryContributorArguments {
    /**
     * Return the first n commits.
     */
    first?: number | null
}

/**
 * A total ref list.
 */
export interface ITotalRefList {
    __typename: 'TotalRefList'

    /**
     * The repositories.
     */
    repositories: IRepository[]

    /**
     * The total.
     */
    total: number
}

/**
 * RepoOrderBy enumerates the ways a repositories-list result set can
 * be ordered.
 */
export enum RepoOrderBy {
    REPO_URI = 'REPO_URI',
    REPO_CREATED_AT = 'REPO_CREATED_AT',
}

/**
 * A list of repositories.
 */
export interface IRepositoryConnection {
    __typename: 'RepositoryConnection'

    /**
     * A list of repositories.
     */
    nodes: IRepository[]

    /**
     *  The total count of repositories in the connection. This total count may be larger
     *  than the number of nodes in this object when the result is paginated.
     *
     *  In some cases, the total count can't be computed quickly; if so, it is null. Pass
     *  precise: true to always compute total counts even if it takes a while.
     */
    totalCount: number | null

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

export interface ITotalCountOnRepositoryConnectionArguments {
    /**
     * @default false
     */
    precise?: boolean | null
}

/**
 * A Phabricator repository.
 */
export interface IPhabricatorRepo {
    __typename: 'PhabricatorRepo'

    /**
     * The canonical repo path (e.g. "github.com/gorilla/mux").
     */
    name: string

    /**
     * An alias for name.
     * @deprecated use name instead
     */
    uri: string

    /**
     * The unique Phabricator identifier for the repo, like "MUX"
     */
    callsign: string

    /**
     * The URL to the phabricator instance (e.g. http://phabricator.sgdev.org)
     */
    url: string
}

/**
 * A period of time in which a set of users have been active.
 */
export enum UserActivePeriod {
    /**
     * Since today at 00:00 UTC.
     */
    TODAY = 'TODAY',

    /**
     * Since the latest Monday at 00:00 UTC.
     */
    THIS_WEEK = 'THIS_WEEK',

    /**
     * Since the first day of the current month at 00:00 UTC.
     */
    THIS_MONTH = 'THIS_MONTH',

    /**
     * All time.
     */
    ALL_TIME = 'ALL_TIME',
}

/**
 * A list of discussion threads.
 */
export interface IDiscussionThreadConnection {
    __typename: 'DiscussionThreadConnection'

    /**
     * A list of discussion threads.
     */
    nodes: IDiscussionThread[]

    /**
     * The total count of discussion threads in the connection. This total
     * count may be larger than the number of nodes in this object when the
     * result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A discussion thread around some target (e.g. a file in a repo).
 */
export interface IDiscussionThread {
    __typename: 'DiscussionThread'

    /**
     * The discussion thread ID (globally unique).
     */
    id: ID

    /**
     * The user who authored this discussion thread.
     */
    author: IUser

    /**
     *  The title of the thread.
     *
     *  Note: the contents of the thread (its 'body') is always the first comment
     *  in the thread. It is always present, even if the user e.g. input no content.
     */
    title: string

    /**
     * The target of this discussion thread.
     */
    target: DiscussionThreadTarget

    /**
     * The date when the discussion thread was created.
     */
    createdAt: string

    /**
     * The date when the discussion thread was last updated.
     */
    updatedAt: string

    /**
     * The date when the discussion thread was archived (or null if it has not).
     */
    archivedAt: string | null

    /**
     * The comments in the discussion thread.
     */
    comments: IDiscussionCommentConnection
}

export interface ICommentsOnDiscussionThreadArguments {
    /**
     * Returns the first n comments from the list.
     */
    first?: number | null
}

/**
 * The target of a discussion thread. Today, the only possible target is a
 * repository. In the future, this may be extended to include other targets such
 * as user profiles, extensions, etc. Clients should ignore target types they
 * do not understand gracefully.
 */
export type DiscussionThreadTarget = IDiscussionThreadTargetRepo

/**
 *  A discussion thread that is centered around:
 *
 *  - A repository.
 *  - A directory inside a repository.
 *  - A file inside a repository.
 *  - A selection inside a file inside a repository.
 *
 */
export interface IDiscussionThreadTargetRepo {
    __typename: 'DiscussionThreadTargetRepo'

    /**
     * The repository in which the thread was created.
     */
    repository: IRepository

    /**
     * The path (relative to the repository root) of the file or directory that
     * the thread is referencing, if any. If the path is null, the thread is not
     * talking about a specific path but rather just the repository generally.
     */
    path: string | null

    /**
     * The branch (but not exact revision) that the thread was referencing, if
     * any.
     */
    branch: IGitRef | null

    /**
     * The exact revision that the thread was referencing, if any.
     */
    revision: IGitRef | null

    /**
     * The selection that the thread was referencing, if any.
     */
    selection: IDiscussionThreadTargetRepoSelection | null
}

/**
 * A selection within a file.
 */
export interface IDiscussionThreadTargetRepoSelection {
    __typename: 'DiscussionThreadTargetRepoSelection'

    /**
     * The line that the selection started on (zero-based, inclusive).
     */
    startLine: number

    /**
     * The character (not byte) of the start line that the selection began on (zero-based, inclusive).
     */
    startCharacter: number

    /**
     * The line that the selection ends on (zero-based, inclusive).
     */
    endLine: number

    /**
     * The character (not byte) of the end line that the selection ended on (zero-based, exclusive).
     */
    endCharacter: number

    /**
     *  The literal textual (UTF-8) lines before the line the selection started
     *  on.
     *
     *  This is an arbitrary number of lines, and may be zero lines, but typically 3.
     */
    linesBefore: string

    /**
     * The literal textual (UTF-8) lines of the selection. i.e. all lines
     * between and including startLine and endLine.
     */
    lines: string

    /**
     *  The literal textual (UTF-8) lines after the line the selection ended on.
     *
     *  This is an arbitrary number of lines, and may be zero lines, but typically 3.
     */
    linesAfter: string
}

/**
 * A list of discussion comments.
 */
export interface IDiscussionCommentConnection {
    __typename: 'DiscussionCommentConnection'

    /**
     * A list of discussion comments.
     */
    nodes: IDiscussionComment[]

    /**
     * The total count of discussion comments in the connection. This total
     * count may be larger than the number of nodes in this object when the
     * result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A comment made within a discussion thread.
 */
export interface IDiscussionComment {
    __typename: 'DiscussionComment'

    /**
     * The discussion comment ID (globally unique).
     */
    id: ID

    /**
     * The discussion thread the comment was made in.
     */
    thread: IDiscussionThread

    /**
     * The user who authored this discussion thread.
     */
    author: IUser

    /**
     *  The actual markdown contents of the comment.
     *
     *  Empty comments are allowed, so that users can create a discussion thread
     *  with a title and no comment contents. Thus, clients should be prepared to
     *  render them. Suggested rendering when the string contains only whitespace
     *  is "_(no comment text)_".
     */
    contents: string

    /**
     *  The markdown contents rendered as an HTML string. It is already sanitized
     *  and escaped and thus is always safe to render.
     *
     *  Empty comments are allowed, so that users can create a discussion thread
     *  with a title and no comment contents. Thus, clients should be prepared to
     *  render them. Suggested rendering when the string contains only whitespace
     *  is "<em>(no comment text)</em>".
     */
    html: string

    /**
     * The date when the discussion thread was created.
     */
    createdAt: string

    /**
     * The date when the discussion thread was last updated.
     */
    updatedAt: string
}

export interface IHtmlOnDiscussionCommentArguments {
    options?: IMarkdownOptions | null
}

/**
 * Describes options for rendering Markdown.
 */
export interface IMarkdownOptions {
    /**
     * A dummy null value (empty input types are not allowed yet).
     */
    alwaysNil?: string | null
}

/**
 * A search.
 */
export interface ISearch {
    __typename: 'Search'

    /**
     * The results.
     */
    results: ISearchResults

    /**
     * The suggestions.
     */
    suggestions: SearchSuggestion[]

    /**
     * A subset of results (excluding actual search results) which are heavily
     * cached and thus quicker to query. Useful for e.g. querying sparkline
     * data.
     */
    stats: ISearchResultsStats
}

export interface ISuggestionsOnSearchArguments {
    first?: number | null
}

/**
 * Search results.
 */
export interface ISearchResults {
    __typename: 'SearchResults'

    /**
     * The results. Inside each SearchResult there may be multiple matches, e.g.
     * a FileMatch may contain multiple line matches.
     */
    results: SearchResult[]

    /**
     *  The total number of results, taking into account the SearchResult type.
     *  This is different than the length of the results array in that e.g. the
     *  results array may contain two file matches and this resultCount would
     *  report 6 ("3 line matches per file").
     *
     *  Typically, 'approximateResultCount', not this field, is shown to users.
     */
    resultCount: number

    /**
     *  The approximate number of results. This is like the length of the results
     *  array, except it can indicate the number of results regardless of whether
     *  or not the limit was hit. Currently, this is represented as e.g. "5+"
     *  results.
     *
     *  This string is typically shown to users to indicate the true result count.
     */
    approximateResultCount: string

    /**
     * Whether or not the results limit was hit.
     */
    limitHit: boolean

    /**
     * Integers representing the sparkline for the search results.
     */
    sparkline: number[]

    /**
     * Repositories that were eligible to be searched.
     */
    repositories: IRepository[]

    /**
     * Repositories that were actually searched. Excludes repositories that would have been searched but were not
     * because a timeout or error occurred while performing the search, or because the result limit was already
     * reached.
     */
    repositoriesSearched: IRepository[]

    /**
     * Indexed repositories searched. This is a subset of repositoriesSearched.
     */
    indexedRepositoriesSearched: IRepository[]

    /**
     * Repositories that are busy cloning onto gitserver.
     */
    cloning: IRepository[]

    /**
     * Repositories or commits that do not exist.
     */
    missing: IRepository[]

    /**
     * Repositories or commits which we did not manage to search in time. Trying
     * again usually will work.
     */
    timedout: IRepository[]

    /**
     * True if indexed search is enabled but was not available during this search.
     */
    indexUnavailable: boolean

    /**
     * An alert message that should be displayed before any results.
     */
    alert: ISearchAlert | null

    /**
     * The time it took to generate these results.
     */
    elapsedMilliseconds: number

    /**
     * Dynamic filters generated by the search results
     */
    dynamicFilters: ISearchFilter[]
}

/**
 * A search result.
 */
export type SearchResult = IFileMatch | ICommitSearchResult | IRepository

/**
 * A file match.
 */
export interface IFileMatch {
    __typename: 'FileMatch'

    /**
     *  The file containing the match.
     *
     *  KNOWN ISSUE: This file's "commit" field contains incomplete data.
     *
     *  KNOWN ISSUE: This field's type should be File! not GitBlob!.
     */
    file: IGitBlob

    /**
     * The repository containing the file match.
     */
    repository: IRepository

    /**
     * The resource.
     * @deprecated use the file field instead
     */
    resource: string

    /**
     * The symbols found in this file that match the query.
     */
    symbols: ISymbol[]

    /**
     * The line matches.
     */
    lineMatches: ILineMatch[]

    /**
     * Whether or not the limit was hit.
     */
    limitHit: boolean
}

/**
 * A line match.
 */
export interface ILineMatch {
    __typename: 'LineMatch'

    /**
     * The preview.
     */
    preview: string

    /**
     * The line number.
     */
    lineNumber: number

    /**
     * Tuples of [offset, length] measured in characters (not bytes).
     */
    offsetAndLengths: number[][]

    /**
     * Whether or not the limit was hit.
     */
    limitHit: boolean
}

/**
 * A search result that is a Git commit.
 */
export interface ICommitSearchResult {
    __typename: 'CommitSearchResult'

    /**
     * The commit that matched the search query.
     */
    commit: IGitCommit

    /**
     * The ref names of the commit.
     */
    refs: IGitRef[]

    /**
     * The refs by which this commit was reached.
     */
    sourceRefs: IGitRef[]

    /**
     * The matching portion of the commit message, if any.
     */
    messagePreview: IHighlightedString | null

    /**
     * The matching portion of the diff, if any.
     */
    diffPreview: IHighlightedString | null
}

/**
 * A string that has highlights (e.g, query matches).
 */
export interface IHighlightedString {
    __typename: 'HighlightedString'

    /**
     * The full contents of the string.
     */
    value: string

    /**
     * Highlighted matches of the query in the preview string.
     */
    highlights: IHighlight[]
}

/**
 * A highlighted region in a string (e.g., matched by a query).
 */
export interface IHighlight {
    __typename: 'Highlight'

    /**
     * The 1-indexed line number.
     */
    line: number

    /**
     * The 1-indexed character on the line.
     */
    character: number

    /**
     * The length of the highlight, in characters (on the same line).
     */
    length: number
}

/**
 * A search-related alert message.
 */
export interface ISearchAlert {
    __typename: 'SearchAlert'

    /**
     * The title.
     */
    title: string

    /**
     * The description.
     */
    description: string | null

    /**
     * "Did you mean: ____" query proposals
     */
    proposedQueries: ISearchQueryDescription[]
}

/**
 * A search query description.
 */
export interface ISearchQueryDescription {
    __typename: 'SearchQueryDescription'

    /**
     * The description.
     */
    description: string | null

    /**
     * The query.
     */
    query: string
}

/**
 * A search filter.
 */
export interface ISearchFilter {
    __typename: 'SearchFilter'

    /**
     * The value.
     */
    value: string

    /**
     * The string to be displayed in the UI.
     */
    label: string

    /**
     * Number of matches for a given filter.
     */
    count: number

    /**
     * Whether the results returned are incomplete.
     */
    limitHit: boolean

    /**
     * The kind of filter. Should be "file" or "repo".
     */
    kind: string
}

/**
 * A search suggestion.
 */
export type SearchSuggestion = IRepository | IFile | ISymbol

/**
 * Statistics about search results.
 */
export interface ISearchResultsStats {
    __typename: 'SearchResultsStats'

    /**
     * The approximate number of results returned.
     */
    approximateResultCount: string

    /**
     * The sparkline.
     */
    sparkline: number[]
}

/**
 * A search scope.
 */
export interface ISearchScope {
    __typename: 'SearchScope'

    /**
     * A unique identifier for the search scope.
     * If set, a scoped search page is available at https://[sourcegraph-hostname]/search/scope/ID, where ID is this value.
     */
    id: string | null

    /**
     * The name.
     */
    name: string

    /**
     * The value.
     */
    value: string

    /**
     * A description for this search scope, which will appear on the scoped search page.
     */
    description: string | null
}

/**
 * A saved search query, defined in configuration.
 */
export interface ISavedQuery {
    __typename: 'SavedQuery'

    /**
     * The unique ID of the saved query.
     */
    id: ID

    /**
     * The subject whose configuration this saved query was defined in.
     */
    subject: ConfigurationSubject

    /**
     * The unique key of this saved query (unique only among all other saved
     * queries of the same subject).
     */
    key: string | null

    /**
     * The 0-indexed index of this saved query in the subject's configuration.
     */
    index: number

    /**
     * The description.
     */
    description: string

    /**
     * The query.
     */
    query: string

    /**
     * Whether or not to show on the homepage.
     */
    showOnHomepage: boolean

    /**
     * Whether or not to notify.
     */
    notify: boolean

    /**
     * Whether or not to notify on Slack.
     */
    notifySlack: boolean
}

/**
 * A group of repositories.
 */
export interface IRepoGroup {
    __typename: 'RepoGroup'

    /**
     * The name.
     */
    name: string

    /**
     * The repositories.
     */
    repositories: string[]
}

/**
 *  A site is an installation of Sourcegraph that consists of one or more
 *  servers that share the same configuration and database.
 *
 *  The site is a singleton; the API only ever returns the single global site.
 */
export interface ISite {
    __typename: 'Site'

    /**
     * The site's opaque GraphQL ID. This is NOT the "site ID" as it is referred to elsewhere;
     * use the siteID field for that. (GraphQL node types conventionally have an id field of type
     * ID! that globally identifies the node.)
     */
    id: ID

    /**
     * The site ID.
     */
    siteID: string

    /**
     * The site's configuration. Only visible to site admins.
     */
    configuration: ISiteConfiguration

    /**
     * The site's latest site-wide settings (which are the lowest-precedence
     * in the configuration cascade for a user).
     */
    latestSettings: ISettings | null

    /**
     * Deprecated settings specified in the site configuration "settings" field. These are distinct from a site's
     * latestSettings (which are stored in the DB) and are applied at the lowest level of precedence.
     */
    deprecatedSiteConfigurationSettings: string | null

    /**
     * The configuration cascade including this subject and all applicable subjects whose configuration is lower
     * precedence than this subject.
     */
    configurationCascade: IConfigurationCascade

    /**
     * The URL to the site's settings.
     */
    settingsURL: string

    /**
     * Whether the viewer can reload the site (with the reloadSite mutation).
     */
    canReloadSite: boolean

    /**
     * Whether the viewer can modify the subject's configuration.
     */
    viewerCanAdminister: boolean

    /**
     * Lists all language servers.
     */
    langServers: ILangServer[]

    /**
     * The language server for a given language (if exists, otherwise null)
     */
    langServer: ILangServer | null

    /**
     *  The status of language server management capabilities.
     *
     *  Only site admins may view this field.
     */
    languageServerManagementStatus: ILanguageServerManagementStatus | null

    /**
     * A list of all access tokens on this site.
     */
    accessTokens: IAccessTokenConnection

    /**
     * A list of all authentication providers.
     */
    authProviders: IAuthProviderConnection

    /**
     * A list of all user external accounts on this site.
     */
    externalAccounts: IExternalAccountConnection

    /**
     * The name of the Sourcegraph product that is used on this site ("Sourcegraph Server" or "Sourcegraph Data
     * Center" when running in production).
     */
    productName: string

    /**
     * The build version of the Sourcegraph software that is running on this site (of the form
     * NNNNN_YYYY-MM-DD_XXXXX, like 12345_2018-01-01_abcdef).
     */
    buildVersion: string

    /**
     * The product version of the Sourcegraph software that is running on this site.
     */
    productVersion: string

    /**
     * Information about software updates for the version of Sourcegraph that this site is running.
     */
    updateCheck: IUpdateCheck

    /**
     * Whether the site needs to be configured to add repositories.
     */
    needsRepositoryConfiguration: boolean

    /**
     * Whether the site has zero access-enabled repositories.
     */
    noRepositoriesEnabled: boolean

    /**
     * Whether the site configuration has validation problems or deprecation notices.
     */
    configurationNotice: boolean

    /**
     * Whether the site has code intelligence. This field will be expanded in the future to describe
     * more about the code intelligence available (languages supported, etc.). It is subject to
     * change without notice.
     */
    hasCodeIntelligence: boolean

    /**
     * Whether the site is using an external authentication service such as OIDC or SAML.
     */
    externalAuthEnabled: boolean

    /**
     * Whether we want to show built-in searches on the saved searches page
     */
    disableBuiltInSearches: boolean

    /**
     * Whether the server sends emails to users to verify email addresses. If false, then site admins must manually
     * verify users' email addresses.
     */
    sendsEmailVerificationEmails: boolean

    /**
     * Information about this site's license to use Sourcegraph software. This is about the license
     * for the use of Sourcegraph itself; it is not about repository licenses or open-source
     * licenses.
     */
    sourcegraphLicense: ISourcegraphLicense

    /**
     * The activity.
     */
    activity: ISiteActivity
}

export interface ILangServerOnSiteArguments {
    language: string
}

export interface IAccessTokensOnSiteArguments {
    /**
     * Returns the first n access tokens from the list.
     */
    first?: number | null
}

export interface IExternalAccountsOnSiteArguments {
    /**
     * Returns the first n external accounts from the list.
     */
    first?: number | null

    /**
     * Include only external accounts associated with this user.
     */
    user?: ID | null

    /**
     * Include only external accounts with this service type.
     */
    serviceType?: string | null

    /**
     * Include only external accounts with this service ID.
     */
    serviceID?: string | null

    /**
     * Include only external accounts with this client ID.
     */
    clientID?: string | null
}

export interface IActivityOnSiteArguments {
    /**
     * Days of history.
     */
    days?: number | null

    /**
     * Weeks of history.
     */
    weeks?: number | null

    /**
     * Months of history.
     */
    months?: number | null
}

/**
 * The configuration for a site.
 */
export interface ISiteConfiguration {
    __typename: 'SiteConfiguration'

    /**
     * The effective configuration JSON. This will lag behind the pendingContents
     * if the site configuration was updated but the server has not yet restarted.
     */
    effectiveContents: string

    /**
     * The pending configuration JSON, which will become effective after the next
     * server restart. This is set if the site configuration has been updated since
     * the server started.
     */
    pendingContents: string | null

    /**
     * Messages describing validation problems or usage of deprecated configuration in the configuration JSON
     * (pendingContents if it exists, otherwise effectiveContents). This includes both JSON Schema validation
     * problems and other messages that perform more advanced checks on the configuration (that can't be expressed
     * in the JSON Schema).
     */
    validationMessages: string[]

    /**
     * Whether the viewer can update the site configuration (using the
     * updateSiteConfiguration mutation).
     */
    canUpdate: boolean

    /**
     * The source of the configuration as a human-readable description,
     * referring to either the on-disk file path or the SOURCEGRAPH_CONFIG
     * env var.
     */
    source: string
}

/**
 * A language server.
 */
export interface ILangServer {
    __typename: 'LangServer'

    /**
     * "go", "java", "typescript", etc.
     */
    language: string

    /**
     * "Go", "Java", "TypeScript", "PHP", etc.
     */
    displayName: string

    /**
     *  Whether or not this language server should be considered experimental.
     *
     *  Has no effect on behavior, only effects how the language server is presented e.g. in the UI.
     */
    experimental: boolean

    /**
     * URL to the language server's homepage, if available.
     */
    homepageURL: string | null

    /**
     * URL to the language server's open/known issues, if available.
     */
    issuesURL: string | null

    /**
     * URL to the language server's documentation, if available.
     */
    docsURL: string | null

    /**
     * Whether or not we are running in Data Center mode.
     */
    dataCenter: boolean

    /**
     * Whether or not this is a custom language server (i.e. one that does not
     * come built in with Sourcegraph).
     */
    custom: boolean

    /**
     *  The current configuration state of the language server.
     *
     *  For custom language servers, this field is never LANG_SERVER_STATE_NONE.
     */
    state: LangServerState

    /**
     *  Whether or not the language server is being downloaded, starting, restarting.
     *
     *  Always false in Data Center and for custom language servers.
     */
    pending: boolean

    /**
     *  Whether or not the language server is being downloaded.
     *
     *  Always false in Data Center and for custom language servers.
     */
    downloading: boolean

    /**
     *  Whether or not the current user can enable the language server or not.
     *
     *  Always false in Data Center.
     */
    canEnable: boolean

    /**
     *  Whether or not the current user can disable the language server or not.
     *
     *  Always false in Data Center.
     */
    canDisable: boolean

    /**
     *  Whether or not the current user can restart the language server or not.
     *
     *  Always false in Data Center and for custom language servers.
     */
    canRestart: boolean

    /**
     *  Whether or not the current user can update the language server or not.
     *
     *  Always false in Data Center and for custom language servers.
     */
    canUpdate: boolean

    /**
     *  Indicates whether or not the language server is healthy or
     *  unhealthy. Examples include:
     *
     *    Healthy:
     *        - Server is running, experiencing no issues.
     *        - Server is not running, currently being downloaded.
     *        - Server is not running, currently starting or restarting.
     *
     *    Unhealthy:
     *        - Server is running, experiencing restarts / OOMs often.
     *        - Server is not running, an error is preventing startup.
     *
     *  The value is true ("healthy") if the language server is not enabled.
     *
     *  Always false in Data Center and for custom language servers.
     */
    healthy: boolean
}

/**
 * The possible configuration states of a language server.
 */
export enum LangServerState {
    /**
     * The language server is neither enabled nor disabled. When a repo for this
     * language is visited by any user, it will be enabled.
     */
    LANG_SERVER_STATE_NONE = 'LANG_SERVER_STATE_NONE',

    /**
     * The language server was enabled by a plain user or admin user.
     */
    LANG_SERVER_STATE_ENABLED = 'LANG_SERVER_STATE_ENABLED',

    /**
     * The language server was disabled by an admin user.
     */
    LANG_SERVER_STATE_DISABLED = 'LANG_SERVER_STATE_DISABLED',
}

/**
 * Status about management capabilities for language servers.
 */
export interface ILanguageServerManagementStatus {
    __typename: 'LanguageServerManagementStatus'

    /**
     *  Whether this site can manage (enable/disable/restart/update) language servers on its own.
     *
     *  Even if this field's value is true, individual language servers may not be manageable. Clients must check the
     *  LangServer.canXyz fields.
     *
     *  Always false on Data Center.
     */
    siteCanManage: boolean

    /**
     * The reason why the site can't manage language servers, if siteCanManage == false.
     */
    reason: string | null
}

/**
 * A list of authentication providers.
 */
export interface IAuthProviderConnection {
    __typename: 'AuthProviderConnection'

    /**
     * A list of authentication providers.
     */
    nodes: IAuthProvider[]

    /**
     * The total count of authentication providers in the connection. This total count may be larger than the number of nodes
     * in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A provider of user authentication, such as an external single-sign-on service (e.g., using OpenID
 * Connect or SAML).
 */
export interface IAuthProvider {
    __typename: 'AuthProvider'

    /**
     * The type of the auth provider.
     */
    serviceType: string

    /**
     * An identifier for the service that the auth provider represents.
     */
    serviceID: string

    /**
     * An identifier for the client of the service that the auth provider represents.
     */
    clientID: string

    /**
     * The human-readable name of the provider.
     */
    displayName: string

    /**
     * Whether this auth provider is the builtin username-password auth provider.
     */
    isBuiltin: boolean

    /**
     * A URL that, when visited, initiates the authentication process for this auth provider.
     */
    authenticationURL: string | null
}

/**
 * Information about software updates for Sourcegraph.
 */
export interface IUpdateCheck {
    __typename: 'UpdateCheck'

    /**
     * Whether an update check is currently in progress.
     */
    pending: boolean

    /**
     * When the last update check was completed, or null if no update check has
     * been completed (or performed) yet.
     */
    checkedAt: string | null

    /**
     * If an error occurred during the last update check, this message describes
     * the error.
     */
    errorMessage: string | null

    /**
     * If an update is available, the version string of the updated version.
     */
    updateVersionAvailable: string | null
}

/**
 * Information about this site's license to use Sourcegraph software. This is about a license for the
 * use of Sourcegraph itself; it is not about a repository license or an open-source license.
 */
export interface ISourcegraphLicense {
    __typename: 'SourcegraphLicense'

    /**
     * An identifier for this Sourcegraph site, generated randomly upon initialization. This value
     * can be overridden by the site admin.
     */
    siteID: string

    /**
     * An email address of the initial site admin.
     */
    primarySiteAdminEmail: string

    /**
     * The total number of users on this Sourcegraph site.
     */
    userCount: number

    /**
     * The Sourcegraph product name ("Sourcegraph Server" or "Sourcegraph Data Center" when running
     * in production).
     */
    productName: string

    /**
     * A list of premium Sourcegraph features and associated information.
     */
    premiumFeatures: ISourcegraphFeature[]
}

/**
 * A feature of Sourcegraph software and associated information.
 */
export interface ISourcegraphFeature {
    __typename: 'SourcegraphFeature'

    /**
     * The title of this feature.
     */
    title: string

    /**
     * A description of this feature.
     */
    description: string

    /**
     * Whether this feature is enabled on this Sourcegraph site.
     */
    enabled: boolean

    /**
     * A URL with more information about this feature.
     */
    informationURL: string
}

/**
 * SiteActivity describes a site's aggregate activity level.
 */
export interface ISiteActivity {
    __typename: 'SiteActivity'

    /**
     * Recent daily active users.
     */
    daus: ISiteActivityPeriod[]

    /**
     * Recent weekly active users.
     */
    waus: ISiteActivityPeriod[]

    /**
     * Recent monthly active users.
     */
    maus: ISiteActivityPeriod[]
}

/**
 * SiteActivityPeriod describes a site's activity level for a given timespan.
 */
export interface ISiteActivityPeriod {
    __typename: 'SiteActivityPeriod'

    /**
     * The time when this started.
     */
    startTime: string

    /**
     * The user count.
     */
    userCount: number

    /**
     * The registered user count.
     */
    registeredUserCount: number

    /**
     * The anonymous user count.
     */
    anonymousUserCount: number
}

/**
 * A list of survey responses
 */
export interface ISurveyResponseConnection {
    __typename: 'SurveyResponseConnection'

    /**
     * A list of survey responses.
     */
    nodes: ISurveyResponse[]

    /**
     * The total count of survey responses in the connection. This total count may be larger
     * than the number of nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * The count of survey responses submitted since 30 calendar days ago at 00:00 UTC.
     */
    last30DaysCount: number

    /**
     * The average score of survey responses in the connection submitted since 30 calendar days ago at 00:00 UTC.
     */
    averageScore: number

    /**
     *  The net promoter score (NPS) of survey responses in the connection submitted since 30 calendar days ago at 00:00 UTC.
     *  Return value is a signed integer, scaled from -100 (all detractors) to +100 (all promoters).
     *
     *  See https://en.wikipedia.org/wiki/Net_Promoter for explanation.
     */
    netPromoterScore: number
}

/**
 * An extension registry.
 */
export interface IExtensionRegistry {
    __typename: 'ExtensionRegistry'

    /**
     *  Find an extension by its extension ID (which is the concatenation of the publisher name, a slash ("/"), and the
     *  extension name).
     *
     *  To find an extension by its GraphQL ID, use Query.node.
     */
    extension: IRegistryExtension | null

    /**
     * A list of extensions published in the extension registry.
     */
    extensions: IRegistryExtensionConnection

    /**
     * A list of publishers with at least 1 extension in the registry.
     */
    publishers: IRegistryPublisherConnection

    /**
     * A list of publishers that the viewer may publish extensions as.
     */
    viewerPublishers: RegistryPublisher[]

    /**
     *  The extension ID prefix for extensions that are published in the local extension registry. This is the
     *  hostname (and port, if non-default HTTP/HTTPS) of the Sourcegraph "appURL" site configuration property.
     *
     *  It is null if extensions published on this Sourcegraph site do not have an extension ID prefix.
     *
     *  Examples: "sourcegraph.example.com/", "sourcegraph.example.com:1234/"
     */
    localExtensionIDPrefix: string | null
}

export interface IExtensionOnExtensionRegistryArguments {
    extensionID: string
}

export interface IExtensionsOnExtensionRegistryArguments {
    /**
     * Returns the first n extensions from the list.
     */
    first?: number | null

    /**
     * Returns only extensions from this publisher.
     */
    publisher?: ID | null

    /**
     * Returns only extensions matching the query.
     */
    query?: string | null

    /**
     * Include extensions from the local registry.
     * @default true
     */
    local?: boolean | null

    /**
     * Include extensions from remote registries.
     * @default true
     */
    remote?: boolean | null

    /**
     *  Sorts the list of extension results such that the extensions with these IDs are first in the result set.
     *
     *  Typically, the client passes the list of added and enabled extension IDs in this parameter so that the
     *  results include those extensions first (which is typically what the user prefers).
     */
    prioritizeExtensionIDs: string[]
}

export interface IPublishersOnExtensionRegistryArguments {
    /**
     * Return the first n publishers from the list.
     */
    first?: number | null
}

/**
 * A list of publishers of extensions in the registry.
 */
export interface IRegistryPublisherConnection {
    __typename: 'RegistryPublisherConnection'

    /**
     * A list of publishers.
     */
    nodes: RegistryPublisher[]

    /**
     * The total count of publishers in the connection. This total count may be larger than the number of
     * nodes in this object when the result is paginated.
     */
    totalCount: number

    /**
     * Pagination information.
     */
    pageInfo: IPageInfo
}

/**
 * A mutation.
 */
export interface IMutation {
    __typename: 'Mutation'

    /**
     *  Updates the user profile information for the user with the given ID.
     *
     *  Only the user and site admins may perform this mutation.
     */
    updateUser: IEmptyResponse

    /**
     *  Creates an organization. The caller is added as a member of the newly created organization.
     *
     *  Only authenticated users may perform this mutation.
     */
    createOrganization: IOrg

    /**
     *  Updates an organization.
     *
     *  Only site admins and any member of the organization may perform this mutation.
     */
    updateOrganization: IOrg

    /**
     * Deletes an organization. Only site admins may perform this mutation.
     */
    deleteOrganization: IEmptyResponse | null

    /**
     *  Adds a repository on a code host that is already present in the site configuration. The name (which may
     *  consist of one or more path components) of the repository must be recognized by an already configured code
     *  host, or else Sourcegraph won't know how to clone it.
     *
     *  The newly added repository is not enabled (unless the code host's configuration specifies that it should be
     *  enabled). The caller must explicitly enable it with setRepositoryEnabled.
     *
     *  If the repository already exists, it is returned.
     *
     *  To add arbitrary repositories (that don't need to reside on an already configured code host), use the site
     *  configuration "repos.list" property.
     *
     *  As a special case, GitHub.com public repositories may be added by using a name of the form
     *  "github.com/owner/repo". If there is no GitHub personal access token for github.com configured, the site may
     *  experience problems with github.com repositories due to the low default github.com API rate limit (60
     *  requests per hour).
     *
     *  Only site admins may perform this mutation.
     */
    addRepository: IRepository

    /**
     *  Enables or disables a repository. A disabled repository is only
     *  accessible to site admins and never appears in search results.
     *
     *  Only site admins may perform this mutation.
     */
    setRepositoryEnabled: IEmptyResponse | null

    /**
     *  Enables or disables all site repositories.
     *
     *  Only site admins may perform this mutation.
     */
    setAllRepositoriesEnabled: IEmptyResponse | null

    /**
     *  Tests the connection to a mirror repository's original source repository. This is an
     *  expensive and slow operation, so it should only be used for interactive diagnostics.
     *
     *  Only site admins may perform this mutation.
     */
    checkMirrorRepositoryConnection: ICheckMirrorRepositoryConnectionResult

    /**
     *  Schedule the mirror repository to be updated from its original source repository. Updating
     *  occurs automatically, so this should not normally be needed.
     *
     *  Only site admins may perform this mutation.
     */
    updateMirrorRepository: IEmptyResponse

    /**
     *  Schedules all repositories to be updated from their original source repositories. Updating
     *  occurs automatically, so this should not normally be needed.
     *
     *  Only site admins may perform this mutation.
     */
    updateAllMirrorRepositories: IEmptyResponse

    /**
     *  Deletes a repository and all data associated with it, irreversibly.
     *
     *  If the repository was added because it was present in the site configuration (directly,
     *  or because it originated from a configured code host), then it will be re-added during
     *  the next sync. If you intend to make the repository inaccessible to users and not searchable,
     *  use setRepositoryEnabled to disable the repository instead of deleteRepository.
     *
     *  Only site admins may perform this mutation.
     */
    deleteRepository: IEmptyResponse | null

    /**
     *  Creates a new user account.
     *
     *  Only site admins may perform this mutation.
     */
    createUser: ICreateUserResult

    /**
     *  Randomize a user's password so that they need to reset it before they can sign in again.
     *
     *  Only site admins may perform this mutation.
     */
    randomizeUserPassword: IRandomizeUserPasswordResult

    /**
     *  Adds an email address to the user's account. The email address will be marked as unverified until the user
     *  has followed the email verification process.
     *
     *  Only the user and site admins may perform this mutation.
     */
    addUserEmail: IEmptyResponse

    /**
     *  Removes an email address from the user's account.
     *
     *  Only the user and site admins may perform this mutation.
     */
    removeUserEmail: IEmptyResponse

    /**
     *  Manually set the verification status of a user's email, without going through the normal verification process
     *  (of clicking on a link in the email with a verification code).
     *
     *  Only site admins may perform this mutation.
     */
    setUserEmailVerified: IEmptyResponse

    /**
     * Deletes a user account. Only site admins may perform this mutation.
     */
    deleteUser: IEmptyResponse | null

    /**
     * Updates the current user's password. The oldPassword arg must match the user's current password.
     */
    updatePassword: IEmptyResponse | null

    /**
     *  Creates an access token that grants the privileges of the specified user (referred to as the access token's
     *  "subject" user after token creation). The result is the access token value, which the caller is responsible
     *  for storing (it is not accessible by Sourcegraph after creation).
     *
     *  The supported scopes are:
     *
     *  - "user:all": Full control of all resources accessible to the user account.
     *  - "site-admin:sudo": Ability to perform any action as any other user. (Only site admins may create tokens
     *    with this scope.)
     *
     *  Only the user or site admins may perform this mutation.
     */
    createAccessToken: ICreateAccessTokenResult

    /**
     *  Deletes and immediately revokes the specified access token, specified by either its ID or by the token
     *  itself.
     *
     *  Only site admins or the user who owns the token may perform this mutation.
     */
    deleteAccessToken: IEmptyResponse

    /**
     *  Deletes the association between an external account and its Sourcegraph user. It does NOT delete the external
     *  account on the external service where it resides.
     *
     *  Only site admins or the user who is associated with the external account may perform this mutation.
     */
    deleteExternalAccount: IEmptyResponse

    /**
     *  Invite the user with the given username to join the organization. The invited user account must already
     *  exist.
     *
     *  Only site admins and any organization member may perform this mutation.
     */
    inviteUserToOrganization: IInviteUserToOrganizationResult

    /**
     *  Accept or reject an existing organization invitation.
     *
     *  Only the recipient of the invitation may perform this mutation.
     */
    respondToOrganizationInvitation: IEmptyResponse

    /**
     *  Resend the notification about an organization invitation to the recipient.
     *
     *  Only site admins and any member of the organization may perform this mutation.
     */
    resendOrganizationInvitationNotification: IEmptyResponse

    /**
     *  Revoke an existing organization invitation.
     *
     *  If the invitation has been accepted or rejected, it may no longer be revoked. After an
     *  invitation is revoked, the recipient may not accept or reject it. Both cases yield an error.
     *
     *  Only site admins and any member of the organization may perform this mutation.
     */
    revokeOrganizationInvitation: IEmptyResponse

    /**
     *  Immediately add a user as a member to the organization, without sending an invitation email.
     *
     *  Only site admins may perform this mutation. Organization members may use the inviteUserToOrganization
     *  mutation to invite users.
     */
    addUserToOrganization: IEmptyResponse

    /**
     *  Removes a user as a member from an organization.
     *
     *  Only site admins and any member of the organization may perform this mutation.
     */
    removeUserFromOrganization: IEmptyResponse | null

    /**
     * Adds a Phabricator repository to Sourcegraph.
     */
    addPhabricatorRepo: IEmptyResponse | null

    /**
     * Resolves a revision for a given diff from Phabricator.
     */
    resolvePhabricatorDiff: IGitCommit | null

    /**
     * Logs a user event.
     */
    logUserEvent: IEmptyResponse | null

    /**
     *  Sends a test notification for the saved search. Be careful: this will send a notifcation (email and other
     *  types of notifications, if configured) to all subscribers of the saved search, which could be bothersome.
     *
     *  Only subscribers to this saved search may perform this action.
     */
    sendSavedSearchTestNotification: IEmptyResponse | null

    /**
     * All mutations that update configuration settings are under this field.
     */
    configurationMutation: IConfigurationMutation | null

    /**
     * Updates the site configuration. Returns whether or not a restart is
     * needed for the update to be applied.
     */
    updateSiteConfiguration: boolean

    /**
     * Manages language servers.
     */
    langServers: ILangServersMutation | null

    /**
     * Manages discussions.
     */
    discussions: IDiscussionsMutation | null

    /**
     *  Sets whether the user with the specified user ID is a site admin.
     * !
     * ! 🚨 SECURITY: Only trusted users should be given site admin permissions.
     * ! Site admins have full access to the site configuration and other
     * ! sensitive data, and they can perform destructive actions such as
     * ! restarting the site.
     */
    setUserIsSiteAdmin: IEmptyResponse | null

    /**
     * Reloads the site by restarting the server. This is not supported for all deployment
     * types. This may cause downtime.
     */
    reloadSite: IEmptyResponse | null

    /**
     * Submits a user satisfaction (NPS) survey.
     */
    submitSurvey: IEmptyResponse | null

    /**
     * Manages the extension registry.
     */
    extensionRegistry: IExtensionRegistryMutation
}

export interface IUpdateUserOnMutationArguments {
    user: ID
    username?: string | null
    displayName?: string | null
    avatarURL?: string | null
}

export interface ICreateOrganizationOnMutationArguments {
    name: string
    displayName?: string | null
}

export interface IUpdateOrganizationOnMutationArguments {
    id: ID
    displayName?: string | null
}

export interface IDeleteOrganizationOnMutationArguments {
    organization: ID
}

export interface IAddRepositoryOnMutationArguments {
    name: string
}

export interface ISetRepositoryEnabledOnMutationArguments {
    repository: ID
    enabled: boolean
}

export interface ISetAllRepositoriesEnabledOnMutationArguments {
    enabled: boolean
}

export interface ICheckMirrorRepositoryConnectionOnMutationArguments {
    /**
     * The ID of the existing repository whose mirror to check.
     */
    repository?: ID | null

    /**
     * The name of a repository whose mirror to check. If the name is provided, the repository need not be added
     * to the site (but the site configuration must define a code host that knows how to handle the name).
     */
    name?: string | null
}

export interface IUpdateMirrorRepositoryOnMutationArguments {
    /**
     * The mirror repository to update.
     */
    repository: ID
}

export interface IDeleteRepositoryOnMutationArguments {
    repository: ID
}

export interface ICreateUserOnMutationArguments {
    /**
     * The new user's username.
     */
    username: string

    /**
     * The new user's optional email address. If given, it is marked as verified.
     */
    email?: string | null
}

export interface IRandomizeUserPasswordOnMutationArguments {
    user: ID
}

export interface IAddUserEmailOnMutationArguments {
    user: ID
    email: string
}

export interface IRemoveUserEmailOnMutationArguments {
    user: ID
    email: string
}

export interface ISetUserEmailVerifiedOnMutationArguments {
    user: ID
    email: string
    verified: boolean
}

export interface IDeleteUserOnMutationArguments {
    user: ID
}

export interface IUpdatePasswordOnMutationArguments {
    oldPassword: string
    newPassword: string
}

export interface ICreateAccessTokenOnMutationArguments {
    user: ID
    scopes: string[]
    note: string
}

export interface IDeleteAccessTokenOnMutationArguments {
    byID?: ID | null
    byToken?: string | null
}

export interface IDeleteExternalAccountOnMutationArguments {
    externalAccount: ID
}

export interface IInviteUserToOrganizationOnMutationArguments {
    organization: ID
    username: string
}

export interface IRespondToOrganizationInvitationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID

    /**
     * The response to the invitation.
     */
    responseType: OrganizationInvitationResponseType
}

export interface IResendOrganizationInvitationNotificationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID
}

export interface IRevokeOrganizationInvitationOnMutationArguments {
    /**
     * The organization invitation.
     */
    organizationInvitation: ID
}

export interface IAddUserToOrganizationOnMutationArguments {
    organization: ID
    username: string
}

export interface IRemoveUserFromOrganizationOnMutationArguments {
    user: ID
    organization: ID
}

export interface IAddPhabricatorRepoOnMutationArguments {
    /**
     * The callsign, for example "MUX".
     */
    callsign: string

    /**
     * The name, for example "github.com/gorilla/mux".
     */
    name?: string | null

    /**
     * An alias for name. DEPRECATED: use name instead.
     */
    uri?: string | null

    /**
     * The URL to the phabricator instance (e.g. http://phabricator.sgdev.org).
     */
    url: string
}

export interface IResolvePhabricatorDiffOnMutationArguments {
    /**
     * The name of the repository that the diff is based on.
     */
    repoName: string

    /**
     * The ID of the diff on Phabricator.
     */
    diffID: ID

    /**
     * The base revision this diff is based on.
     */
    baseRev: string

    /**
     * The raw contents of the diff from Phabricator.
     * Required if Sourcegraph doesn't have a Conduit API token.
     */
    patch?: string | null

    /**
     * The description of the diff. This will be used as the commit message.
     */
    description?: string | null

    /**
     * The name of author of the diff.
     */
    authorName?: string | null

    /**
     * The author's email.
     */
    authorEmail?: string | null

    /**
     * When the diff was created.
     */
    date?: string | null
}

export interface ILogUserEventOnMutationArguments {
    event: UserEvent
    userCookieID: string
}

export interface ISendSavedSearchTestNotificationOnMutationArguments {
    /**
     * ID of the saved search.
     */
    id: ID
}

export interface IConfigurationMutationOnMutationArguments {
    input: IConfigurationMutationGroupInput
}

export interface IUpdateSiteConfigurationOnMutationArguments {
    input: string
}

export interface ISetUserIsSiteAdminOnMutationArguments {
    userID: ID
    siteAdmin: boolean
}

export interface ISubmitSurveyOnMutationArguments {
    input: ISurveySubmissionInput
}

/**
 * Represents a null return value.
 */
export interface IEmptyResponse {
    __typename: 'EmptyResponse'

    /**
     * A dummy null value.
     */
    alwaysNil: string | null
}

/**
 * The result for Mutation.checkMirrorRepositoryConnection.
 */
export interface ICheckMirrorRepositoryConnectionResult {
    __typename: 'CheckMirrorRepositoryConnectionResult'

    /**
     * The error message encountered during the update operation, if any. If null, then
     * the connection check succeeded.
     */
    error: string | null
}

/**
 * The result for Mutation.createUser.
 */
export interface ICreateUserResult {
    __typename: 'CreateUserResult'

    /**
     * The new user.
     */
    user: IUser

    /**
     * The reset password URL that the new user must visit to sign into their account. If the builtin
     * username-password authentication provider is not enabled, this field's value is null.
     */
    resetPasswordURL: string | null
}

/**
 * The result for Mutation.randomizeUserPassword.
 */
export interface IRandomizeUserPasswordResult {
    __typename: 'RandomizeUserPasswordResult'

    /**
     * The reset password URL that the user must visit to sign into their account again. If the builtin
     * username-password authentication provider is not enabled, this field's value is null.
     */
    resetPasswordURL: string | null
}

/**
 * The result for Mutation.createAccessToken.
 */
export interface ICreateAccessTokenResult {
    __typename: 'CreateAccessTokenResult'

    /**
     * The ID of the newly created access token.
     */
    id: ID

    /**
     * The secret token value that is used to authenticate API clients. The caller is responsible for storing this
     * value.
     */
    token: string
}

/**
 * The result of Mutation.inviteUserToOrganization.
 */
export interface IInviteUserToOrganizationResult {
    __typename: 'InviteUserToOrganizationResult'

    /**
     * Whether an invitation email was sent. If emails are not enabled on this site or if the user has no verified
     * email address, an email will not be sent.
     */
    sentInvitationEmail: boolean

    /**
     * The URL that the invited user can visit to accept or reject the invitation.
     */
    invitationURL: string
}

/**
 * A user event.
 */
export enum UserEvent {
    PAGEVIEW = 'PAGEVIEW',
    SEARCHQUERY = 'SEARCHQUERY',
    CODEINTEL = 'CODEINTEL',
    CODEINTELINTEGRATION = 'CODEINTELINTEGRATION',
}

/**
 * Input for Mutation.configuration, which contains fields that all configuration
 * mutations need.
 */
export interface IConfigurationMutationGroupInput {
    /**
     * The subject whose configuration to mutate (organization, user, etc.).
     */
    subject: ID

    /**
     * The ID of the last-known configuration known to the client, or null if
     * there is none. This field is used to prevent race conditions when there
     * are concurrent editors.
     */
    lastID?: number | null
}

/**
 *  Mutations that update configuration settings. These mutations are grouped
 *  together because they:
 *
 *  - are all versioned to avoid race conditions with concurrent editors
 *  - all apply to a specific configuration subject
 *
 *  Grouping them lets us extract those common parameters to the
 *  Mutation.configuration field.
 */
export interface IConfigurationMutation {
    __typename: 'ConfigurationMutation'

    /**
     * Edit a single property in the configuration object.
     */
    editConfiguration: IUpdateConfigurationPayload | null

    /**
     * Overwrite the contents to the new contents provided.
     */
    overwriteConfiguration: IUpdateConfigurationPayload | null

    /**
     * Create a saved query.
     */
    createSavedQuery: ISavedQuery

    /**
     * Update the saved query with the given ID in the configuration.
     */
    updateSavedQuery: ISavedQuery

    /**
     * Delete the saved query with the given ID in the configuration.
     */
    deleteSavedQuery: IEmptyResponse | null
}

export interface IEditConfigurationOnConfigurationMutationArguments {
    /**
     * The configuration edit to apply.
     */
    edit: IConfigurationEdit
}

export interface IOverwriteConfigurationOnConfigurationMutationArguments {
    contents?: string | null
}

export interface ICreateSavedQueryOnConfigurationMutationArguments {
    description: string
    query: string

    /**
     * @default false
     */
    showOnHomepage?: boolean | null

    /**
     * @default false
     */
    notify?: boolean | null

    /**
     * @default false
     */
    notifySlack?: boolean | null

    /**
     * @default false
     */
    disableSubscriptionNotifications?: boolean | null
}

export interface IUpdateSavedQueryOnConfigurationMutationArguments {
    id: ID
    description?: string | null
    query?: string | null

    /**
     * @default false
     */
    showOnHomepage?: boolean | null

    /**
     * @default false
     */
    notify?: boolean | null

    /**
     * @default false
     */
    notifySlack?: boolean | null
}

export interface IDeleteSavedQueryOnConfigurationMutationArguments {
    id: ID

    /**
     * @default false
     */
    disableSubscriptionNotifications?: boolean | null
}

/**
 * An edit to a (nested) configuration property's value.
 */
export interface IConfigurationEdit {
    /**
     *  The key path of the property to update.
     *
     *  Inserting into an existing array is not yet supported.
     */
    keyPath: IKeyPathSegment[]

    /**
     *  The new JSON-encoded value to insert. If the field's value is not set, the property is removed. (This is
     *  different from the field's value being the JSON null value.)
     *
     *  When the value is a non-primitive type, it must be specified using a GraphQL variable, not an inline literal,
     *  or else the GraphQL parser will return an error.
     */
    value?: any | null

    /**
     * Whether to treat the value as a JSONC-encoded string, which makes it possible to perform a configuration edit
     * that preserves (or adds/removes) comments.
     * @default false
     */
    valueIsJSONCEncodedString?: boolean | null
}

/**
 *  A segment of a key path that locates a nested JSON value in a root JSON value. Exactly one field in each
 *  KeyPathSegment must be non-null.
 *
 *  For example, in {"a": [0, {"b": 3}]}, the value 3 is located at the key path ["a", 1, "b"].
 */
export interface IKeyPathSegment {
    /**
     * The name of the property in the object at this location to descend into.
     */
    property?: string | null

    /**
     * The index of the array at this location to descend into.
     */
    index?: number | null
}

/**
 * The payload for ConfigurationMutation.updateConfiguration.
 */
export interface IUpdateConfigurationPayload {
    __typename: 'UpdateConfigurationPayload'

    /**
     * An empty response.
     */
    empty: IEmptyResponse | null
}

/**
 * Mutations for language servers.
 */
export interface ILangServersMutation {
    __typename: 'LangServersMutation'

    /**
     *  Enables the language server for the given language.
     *
     *  Any user can perform this mutation, unless the language has been
     *  explicitly disabled.
     */
    enable: IEmptyResponse | null

    /**
     *  Disables the language server for the given language.
     *
     *  Only admins can perform this action. After disabling, it is impossible
     *  for plain users to enable the language server for this language (until an
     *  admin re-enables it).
     */
    disable: IEmptyResponse | null

    /**
     *  Restarts the language server for the given language.
     *
     *  Only admins can perform this action.
     */
    restart: IEmptyResponse | null

    /**
     *  Updates the language server for the given language.
     *
     *  Only admins can perform this action.
     */
    update: IEmptyResponse | null
}

export interface IEnableOnLangServersMutationArguments {
    language: string
}

export interface IDisableOnLangServersMutationArguments {
    language: string
}

export interface IRestartOnLangServersMutationArguments {
    language: string
}

export interface IUpdateOnLangServersMutationArguments {
    language: string
}

/**
 * Mutations for discussions.
 */
export interface IDiscussionsMutation {
    __typename: 'DiscussionsMutation'

    /**
     * Creates a new thread. Returns the new thread.
     */
    createThread: IDiscussionThread

    /**
     * Updates an existing thread. Returns the updated thread.
     */
    updateThread: IDiscussionThread

    /**
     * Adds a new comment to a thread. Returns the updated thread.
     */
    addCommentToThread: IDiscussionThread
}

export interface ICreateThreadOnDiscussionsMutationArguments {
    input: IDiscussionThreadCreateInput
}

export interface IUpdateThreadOnDiscussionsMutationArguments {
    input: IDiscussionThreadUpdateInput
}

export interface IAddCommentToThreadOnDiscussionsMutationArguments {
    threadID: ID
    contents: string
}

/**
 * Describes the creation of a new thread around some target (e.g. a file in a repo).
 */
export interface IDiscussionThreadCreateInput {
    /**
     * The title of the thread's first comment (i.e. the threads title).
     */
    title: string

    /**
     * The contents of the thread's first comment (i.e. the threads comment).
     */
    contents: string

    /**
     * The target repo of this discussion thread. This is nullable so that in
     * the future more target types may be added.
     */
    targetRepo?: IDiscussionThreadTargetRepoInput | null
}

/**
 *  A discussion thread that is centered around:
 *
 *  - A repository.
 *  - A directory inside a repository.
 *  - A file inside a repository.
 *  - A selection inside a file inside a repository.
 *
 */
export interface IDiscussionThreadTargetRepoInput {
    /**
     * The repository in which the thread was created.
     */
    repository: ID

    /**
     * The path (relative to the repository root) of the file or directory that
     * the thread is referencing, if any. If the path is null, the thread is not
     * talking about a specific path but rather just the repository generally.
     */
    path?: string | null

    /**
     * The branch (but not exact revision) that the thread was referencing, if
     * any.
     */
    branch?: string | null

    /**
     * The exact revision that the thread was referencing, if any.
     */
    revision?: string | null

    /**
     * The selection that the thread was referencing, if any.
     */
    selection?: IDiscussionThreadTargetRepoSelectionInput | null
}

/**
 * A selection within a file.
 */
export interface IDiscussionThreadTargetRepoSelectionInput {
    /**
     * The line that the selection started on (zero-based, inclusive).
     */
    startLine: number

    /**
     * The character (not byte) of the start line that the selection began on (zero-based, inclusive).
     */
    startCharacter: number

    /**
     * The line that the selection ends on (zero-based, inclusive).
     */
    endLine: number

    /**
     * The character (not byte) of the end line that the selection ended on (zero-based, exclusive).
     */
    endCharacter: number

    /**
     *  The literal textual (UTF-8) lines before the line the selection started
     *  on.
     *
     *  This is an arbitrary number of lines, and may be zero lines, but typically 3.
     */
    linesBefore: string

    /**
     * The literal textual (UTF-8) lines of the selection. i.e. all lines
     * between and including startLine and endLine.
     */
    lines: string

    /**
     *  The literal textual (UTF-8) lines after the line the selection ended on.
     *
     *  This is an arbitrary number of lines, and may be zero lines, but typically 3.
     */
    linesAfter: string
}

/**
 * Describes an update mutation to an existing thread.
 */
export interface IDiscussionThreadUpdateInput {
    /**
     * The ID of the thread to update.
     */
    ThreadID: ID

    /**
     * When non-null, indicates that the thread should be archived.
     */
    Archive?: boolean | null

    /**
     * When non-null, indicates that the thread should be deleted. Only admins
     * can perform this action.
     */
    Delete?: boolean | null
}

/**
 * Input for a user satisfaction (NPS) survey submission.
 */
export interface ISurveySubmissionInput {
    /**
     * User-provided email address, if there is no currently authenticated user. If there is, this value
     * will not be used.
     */
    email?: string | null

    /**
     * User's likelihood of recommending Sourcegraph to a friend, from 0-10.
     */
    score: number

    /**
     * The answer to "What is the most important reason for the score you gave".
     */
    reason?: string | null

    /**
     * The answer to "What can Sourcegraph do to provide a better product"
     */
    better?: string | null
}

/**
 * Mutations for the extension registry.
 */
export interface IExtensionRegistryMutation {
    __typename: 'ExtensionRegistryMutation'

    /**
     * Create a new extension in the extension registry.
     */
    createExtension: IExtensionRegistryCreateExtensionResult

    /**
     *  Update an extension in the extension registry.
     *
     *  Only authorized extension publishers may perform this mutation.
     */
    updateExtension: IExtensionRegistryUpdateExtensionResult

    /**
     *  Delete an extension from the extension registry.
     *
     *  Only authorized extension publishers may perform this mutation.
     */
    deleteExtension: IEmptyResponse
}

export interface ICreateExtensionOnExtensionRegistryMutationArguments {
    /**
     * The ID of the extension's publisher (a user or organization).
     */
    publisher: ID

    /**
     * The name of the extension.
     */
    name: string
}

export interface IUpdateExtensionOnExtensionRegistryMutationArguments {
    /**
     * The extension to update.
     */
    extension: ID

    /**
     * The new name for the extension, or null to leave unchanged.
     */
    name?: string | null

    /**
     * The new manifest for the extension, or null to leave unchanged.
     */
    manifest?: string | null
}

export interface IDeleteExtensionOnExtensionRegistryMutationArguments {
    /**
     * The ID of the extension to delete.
     */
    extension: ID
}

/**
 * The result of Mutation.extensionRegistry.createExtension.
 */
export interface IExtensionRegistryCreateExtensionResult {
    __typename: 'ExtensionRegistryCreateExtensionResult'

    /**
     * The newly created extension.
     */
    extension: IRegistryExtension
}

/**
 * The result of Mutation.extensionRegistry.updateExtension.
 */
export interface IExtensionRegistryUpdateExtensionResult {
    __typename: 'ExtensionRegistryUpdateExtensionResult'

    /**
     * The newly updated extension.
     */
    extension: IRegistryExtension
}

/**
 * A diff between two diffable Git objects.
 */
export interface IDiff {
    __typename: 'Diff'

    /**
     * The diff's repository.
     */
    repository: IRepository

    /**
     * The revision range of the diff.
     */
    range: IGitRevisionRange
}

/**
 * A search result that is a diff between two diffable Git objects.
 */
export interface IDiffSearchResult {
    __typename: 'DiffSearchResult'

    /**
     * The diff that matched the search query.
     */
    diff: IDiff

    /**
     * The matching portion of the diff.
     */
    preview: IHighlightedString
}

/**
 * Ref fields.
 */
export interface IRefFields {
    __typename: 'RefFields'

    /**
     * The ref location.
     */
    refLocation: IRefLocation | null

    /**
     * The URI.
     */
    uri: IURI | null
}

/**
 * A ref location.
 */
export interface IRefLocation {
    __typename: 'RefLocation'

    /**
     * The starting line number.
     */
    startLineNumber: number

    /**
     * The starting column.
     */
    startColumn: number

    /**
     * The ending line number.
     */
    endLineNumber: number

    /**
     * The ending column.
     */
    endColumn: number
}

/**
 * A URI.
 */
export interface IURI {
    __typename: 'URI'

    /**
     * The host.
     */
    host: string

    /**
     * The fragment.
     */
    fragment: string

    /**
     * The path.
     */
    path: string

    /**
     * The query.
     */
    query: string

    /**
     * The scheme.
     */
    scheme: string
}

/**
 * A deployment configuration.
 */
export interface IDeploymentConfiguration {
    __typename: 'DeploymentConfiguration'

    /**
     * The email.
     */
    email: string | null

    /**
     * The site ID.
     */
    siteID: string | null
}
