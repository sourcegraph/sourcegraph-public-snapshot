declare namespace GQL {
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
         * The current site settings.
         */
        currentSiteSettings: ISettings | null

        /**
         * The configuration.
         */
        configuration: IConfigurationCascade

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
         * Looks up an organization by ID.
         * @deprecated use Query.node instead
         */
        org: IOrg

        /**
         * Looks up a shared item by ULID. This is an experimental feature.
         */
        sharedItem: ISharedItem | null

        /**
         * The current site.
         */
        site: ISite

        /**
         * Retrieve responses to surveys.
         */
        surveyResponses: ISurveyResponseConnection
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

    export interface ISearchOnQueryArguments {
        /**
         * The search query (such as "foo" or "repo:myrepo foo").
         * @default
         */
        query?: string | null
    }

    export interface IOrgOnQueryArguments {
        id: ID
    }

    export interface ISharedItemOnQueryArguments {
        ulid: string
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
        | IThread
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
         * Returns information about the given commit in the repository.
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
        rev: string
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
         * The URL to this commit.
         */
        url: string

        /**
         * The URLs to this commit on its repository's external services.
         */
        externalURLs: IExternalLink[]

        /**
         * Lists the Git tree as of this commit.
         */
        tree: ITree | null

        /**
         * Retrieves a Git blob (file) as of this commit.
         */
        file: IFile | null

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
         * @default
         */
        path?: string | null

        /**
         * @default false
         */
        recursive?: boolean | null
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
         * The organizations that this user is a member of.
         */
        orgs: IOrg[]

        /**
         * This user's organization memberships.
         */
        orgMemberships: IOrgMembership[]

        /**
         *  The internal tags associated with the user. This is an internal site management feature.
         *
         *  Only the user and site admins can access this field.
         */
        tags: IUserTag[]

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
         *  The users' access tokens (which grant to the holder the privileges of the user). This consists
         *  of all access tokens whose subject is this user.
         *
         *  Only the user and site admins can access this field.
         */
        accessTokens: IAccessTokenConnection

        /**
         * A list of external accounts that are associated with the user.
         */
        externalAccounts: IExternalAccount[]

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
    }

    export interface IAccessTokensOnUserArguments {
        /**
         * Returns the first n access tokens from the list.
         */
        first?: number | null
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
     * An organization, which is a group of users.
     */
    export interface IOrg {
        __typename: 'Org'

        /**
         * The unique ID for the organization.
         */
        id: ID

        /**
         * The numeric unique ID for the organization.
         * @deprecated use id instead
         */
        orgID: number

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
         *  The repositories associated with the organization. This is an experimental feature.
         *
         *  Only organization members and site admins can access this field.
         */
        repos: IOrgRepo[]

        /**
         *  Look up a repository associated with the organization. This is an experimental feature.
         *
         *  Only organization members and site admins can access this field.
         */
        repo: IOrgRepo | null

        /**
         *  Threads associated with the organization. This is an experimental feature.
         *
         *  Only organization members and site admins can access this field.
         */
        threads: IThreadConnection

        /**
         *  The internal tags associated with the organization. This is an internal site management feature.
         *
         *  Only organization members and site admins can access this field.
         */
        tags: IOrgTag[]

        /**
         * Whether the viewer has admin privileges on this organization. Currently, all of an organization's members
         * have admin privileges on the organization.
         */
        viewerCanAdminister: boolean

        /**
         * Whether the viewer is a member of this organization.
         */
        viewerIsMember: boolean
    }

    export interface IRepoOnOrgArguments {
        canonicalRemoteID: string
    }

    export interface IThreadsOnOrgArguments {
        /**
         * DEPRECATED: use canonicalRemoteIDs instead.
         */
        repoCanonicalRemoteID?: string | null
        canonicalRemoteIDs: string[]
        branch?: string | null
        file?: string | null
        limit?: number | null
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
     * An organization repository.
     */
    export interface IOrgRepo {
        __typename: 'OrgRepo'

        /**
         * The ID.
         */
        id: number

        /**
         * The organization.
         */
        org: IOrg

        /**
         * The canonical remote ID.
         */
        canonicalRemoteID: string

        /**
         * The time when this was created.
         */
        createdAt: string

        /**
         * The time when this was updated.
         */
        updatedAt: string

        /**
         * Gets the threads. This is an experimental feature.
         */
        threads: IThreadConnection

        /**
         * The repository that this refers to, if the repository is available on the server. This is null
         * for repositories that only exist for users locally (that they use with the editor) but that
         * are not on the server.
         */
        repository: IRepository | null
    }

    export interface IThreadsOnOrgRepoArguments {
        file?: string | null
        branch?: string | null
        limit?: number | null
    }

    /**
     * A list of threads.
     */
    export interface IThreadConnection {
        __typename: 'ThreadConnection'

        /**
         * A list of threads.
         */
        nodes: IThread[]

        /**
         * The total count of threads in the connection. This total count may be larger
         * than the number of nodes in this object when the result is paginated.
         */
        totalCount: number
    }

    /**
     * Thread is a comment thread.
     */
    export interface IThread {
        __typename: 'Thread'

        /**
         * The unique ID.
         */
        id: ID

        /**
         * The primary key from the database.
         */
        databaseID: number

        /**
         * The repository.
         */
        repo: IOrgRepo

        /**
         * The file.
         * @deprecated use repoRevisionPath (or linesRevisionPath) instead
         */
        file: string

        /**
         * The relative path of the resource in the repository at repoRevision.
         */
        repoRevisionPath: string

        /**
         * The relative path of the resource in the repository at linesRevision.
         */
        linesRevisionPath: string

        /**
         * The branch.
         */
        branch: string | null

        /**
         * The commit ID of the repository at the time the thread was created.
         */
        repoRevision: string

        /**
         *  The commit ID from Git blame, at the time the thread was created.
         *
         *  The selection may be multiple lines, and the commit id is the
         *  topologically most recent commit of the blame commit ids for the selected
         *  lines.
         *
         *  For example, if you have a selection of lines that have blame revisions
         *  (a, c, e, f), and assuming a history like::
         *
         *  	a <- b <- c <- d <- e <- f <- g <- h <- HEAD
         *
         *  Then lines_revision would be f, because all other blame revisions a, c, e
         *  are reachable from f.
         *
         *  Or in lay terms: "What is the oldest revision that I could checkout and
         *  still see the exact lines of code that I selected?".
         */
        linesRevision: string

        /**
         * The title.
         */
        title: string

        /**
         * The start line.
         */
        startLine: number

        /**
         * The end line.
         */
        endLine: number

        /**
         * The start character.
         */
        startCharacter: number

        /**
         * The end character.
         */
        endCharacter: number

        /**
         * The range length.
         */
        rangeLength: number

        /**
         * The time when this was created.
         */
        createdAt: string

        /**
         * The time when this was archived.
         */
        archivedAt: string | null

        /**
         * The author.
         */
        author: IUser

        /**
         * The lines.
         */
        lines: IThreadLines | null

        /**
         * The comments.
         */
        comments: IComment[]
    }

    /**
     * Thread lines.
     */
    export interface IThreadLines {
        __typename: 'ThreadLines'

        /**
         *  HTML context lines before 'html'.
         *
         *  It is sanitized already by the server, and thus is safe for rendering.
         */
        htmlBefore: string

        /**
         *  HTML lines that the user's selection was made on.
         *
         *  It is sanitized already by the server, and thus is safe for rendering.
         */
        html: string

        /**
         *  HTML context lines after 'html'.
         *
         *  It is sanitized already by the server, and thus is safe for rendering.
         */
        htmlAfter: string

        /**
         * Text context lines before 'text'.
         */
        textBefore: string

        /**
         * Text lines that the user's selection was made on.
         */
        text: string

        /**
         * Text context lines after 'text'.
         */
        textAfter: string

        /**
         * Byte offset into textLines where user selection began.
         */
        textSelectionRangeStart: number

        /**
         * Length in bytes of the user selection.
         */
        textSelectionRangeLength: number
    }

    export interface IHtmlBeforeOnThreadLinesArguments {
        isLightTheme: boolean
    }

    export interface IHtmlOnThreadLinesArguments {
        isLightTheme: boolean
    }

    export interface IHtmlAfterOnThreadLinesArguments {
        isLightTheme: boolean
    }

    /**
     * Comment is a comment in a thread.
     */
    export interface IComment {
        __typename: 'Comment'

        /**
         * The unique ID.
         */
        id: ID

        /**
         * The primary key from the database.
         */
        databaseID: number

        /**
         * The title.
         */
        title: string

        /**
         * The contents.
         */
        contents: string

        /**
         *  The file rendered as rich HTML, or an empty string if it is not a supported
         *  rich file type.
         *
         *  This HTML string is already escaped and thus is always safe to render.
         */
        richHTML: string

        /**
         * The time when his was created.
         */
        createdAt: string

        /**
         * The time when this was updated.
         */
        updatedAt: string

        /**
         * The author.
         */
        author: IUser
    }

    /**
     * An organization tag.
     */
    export interface IOrgTag {
        __typename: 'OrgTag'

        /**
         * The ID.
         */
        id: number

        /**
         * The name.
         */
        name: string
    }

    /**
     * An organization membership.
     */
    export interface IOrgMembership {
        __typename: 'OrgMembership'

        /**
         * The ID.
         */
        id: number

        /**
         * The organization.
         */
        org: IOrg

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
     * A user tag.
     */
    export interface IUserTag {
        __typename: 'UserTag'

        /**
         * The ID.
         */
        id: number

        /**
         * The name.
         */
        name: string
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
     * A tree.
     */
    export interface ITree {
        __typename: 'Tree'

        /**
         * The directories.
         */
        directories: IDirectory[]

        /**
         * The files
         */
        files: IFile[]

        /**
         * Consists of directories plus files.
         */
        entries: TreeEntry[]

        /**
         *  FOR INTERNAL USE ONLY.
         *
         *  An optimized, raw encoding of this tree, used by the Sourcegraph frontend web application's file tree
         *  component.
         */
        internalRaw: string
    }

    /**
     * A directory.
     */
    export interface IDirectory {
        __typename: 'Directory'

        /**
         * The full path (relative to the repository root) of this directory.
         */
        path: string

        /**
         * The base name (i.e., file name only) of this directory.
         */
        name: string

        /**
         * True because this is a directory. (The value differs for other TreeEntry interface implementations, such as
         * File.)
         */
        isDirectory: boolean

        /**
         * The repository containing this directory.
         */
        repository: IRepository

        /**
         * The list of Git commits that touched this directory.
         */
        commits: IGitCommit[]

        /**
         * The URL to this directory.
         */
        url: string

        /**
         * The tree.
         */
        tree: ITree

        /**
         * Symbols defined in this directory.
         */
        symbols: ISymbolConnection
    }

    export interface ISymbolsOnDirectoryArguments {
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
    export type TreeEntry = IDirectory | IFile

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
         * The repository containing this tree entry.
         */
        repository: IRepository

        /**
         * The list of Git commits that touched this tree entry.
         */
        commits: IGitCommit[]

        /**
         * The URL to this tree entry.
         */
        url: string

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
         * The URL of this symbol.
         */
        url: string
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
        resource: TreeEntry

        /**
         * The range inside the file that this location refers to.
         */
        range: IRange | null

        /**
         * The URL to this location.
         */
        url: string
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
     * A file.
     */
    export interface IFile {
        __typename: 'File'

        /**
         * The full path (relative to the repository root) of this file.
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
         * The repository containing this file.
         */
        repository: IRepository

        /**
         * The list of Git commits that touched this file.
         */
        commits: IGitCommit[]

        /**
         * The URL to this file.
         */
        url: string

        /**
         * The content of this file.
         */
        content: string

        /**
         *  The file rendered as rich HTML, or an empty string if it is not a supported
         *  rich file type.
         *
         *  This HTML string is already escaped and thus is always safe to render.
         */
        richHTML: string

        /**
         * The URLs to this file on its repository's external services.
         */
        externalURLs: IExternalLink[]

        /**
         * Whether or not it is binary.
         */
        binary: boolean

        /**
         * Highlight the file.
         */
        highlight: IHighlightedFile

        /**
         * Blame the file.
         */
        blame: IHunk[]

        /**
         * Returns dependency references for the file.
         */
        dependencyReferences: IDependencyReferences

        /**
         * Blame the file and return output in a raw string format.
         */
        blameRaw: string

        /**
         * Symbols defined in this file.
         */
        symbols: ISymbolConnection
    }

    export interface IHighlightOnFileArguments {
        disableTimeout: boolean
        isLightTheme: boolean
    }

    export interface IBlameOnFileArguments {
        startLine: number
        endLine: number
    }

    export interface IDependencyReferencesOnFileArguments {
        Language: string
        Line: number
        Character: number
    }

    export interface IBlameRawOnFileArguments {
        startLine: number
        endLine: number
    }

    export interface ISymbolsOnFileArguments {
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
        author: ISignature | null

        /**
         * The message.
         */
        message: string
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
     * ALl possible types of Git refs.
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
         * The repository containing this file.
         */
        repository: IRepository

        /**
         * The old (original) path of the file, or null if the file was added.
         */
        oldPath: string | null

        /**
         * The new (changed) path of the file, or null if the file was deleted.
         */
        newPath: string | null

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
     * The configurations for all of the relevant configuration subjects, plus the merged
     * configuration.
     */
    export interface IConfigurationCascade {
        __typename: 'ConfigurationCascade'

        /**
         * The default settings, which are applied first and the lowest priority behind
         * all configuration subjects' settings.
         */
        defaults: IConfiguration | null

        /**
         * The configurations for all of the subjects that are applied for the currently
         * authenticated user. For example, a user in 2 orgs would have the following
         * configuration subjects: org 1, org 2, and the user.
         */
        subjects: ConfigurationSubject[]

        /**
         * The effective configuration, merged from all of the subjects.
         */
        merged: IConfiguration
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
         * The results.
         */
        results: SearchResult[]

        /**
         * The number of results returned.
         */
        resultCount: number

        /**
         * The approximate number of results returned.
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
        repositories: string[]

        /**
         * Repositories that were actually searched. Excludes repositories that would have been searched but were not
         * because a timeout or error occurred while performing the search, or because the result limit was already
         * reached.
         */
        repositoriesSearched: string[]

        /**
         * Indexed repositories searched. This is a subset of repositoriesSearched.
         */
        indexedRepositoriesSearched: string[]

        /**
         * Repositories that are busy cloning onto gitserver.
         */
        cloning: string[]

        /**
         * Repositories or commits that do not exist.
         */
        missing: string[]

        /**
         * Repositories or commits which we did not manage to search in time. Trying
         * again usually will work.
         */
        timedout: string[]

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
         * The resource.
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
     *  Represents a shared item (either a shared code comment OR code snippet).
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItem {
        __typename: 'SharedItem'

        /**
         * Who shared the item.
         */
        author: ISharedItemUser

        /**
         * Whether or not this shared item is public.
         */
        public: boolean

        /**
         * The thread.
         */
        thread: ISharedItemThread

        /**
         * Present only if the shared item was a specific comment.
         */
        comment: ISharedItemComment | null
    }

    /**
     *  Like the User type, except with fields that should not be accessible with a
     *  secret URL removed.
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItemUser {
        __typename: 'SharedItemUser'

        /**
         * The display name.
         */
        displayName: string | null

        /**
         * The username.
         */
        username: string

        /**
         * The avatar URL.
         */
        avatarURL: string | null
    }

    /**
     *  Like the Thread type, except with fields that should not be accessible with a
     *  secret URL removed.
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItemThread {
        __typename: 'SharedItemThread'

        /**
         * The id.
         */
        id: ID

        /**
         * The database ID.
         */
        databaseID: number

        /**
         * The repository.
         */
        repo: ISharedItemOrgRepo

        /**
         * The file.
         */
        file: string

        /**
         * The branch.
         */
        branch: string | null

        /**
         * The repository evision.
         */
        repoRevision: string

        /**
         * The lines revision.
         */
        linesRevision: string

        /**
         * The title.
         */
        title: string

        /**
         * The start line.
         */
        startLine: number

        /**
         * The end line.
         */
        endLine: number

        /**
         * The start character.
         */
        startCharacter: number

        /**
         * The end character.
         */
        endCharacter: number

        /**
         * The range length.
         */
        rangeLength: number

        /**
         * The time when this was created.
         */
        createdAt: string

        /**
         * The time when this was archived.
         */
        archivedAt: string | null

        /**
         * The author.
         */
        author: ISharedItemUser

        /**
         * The lines.
         */
        lines: ISharedItemThreadLines | null

        /**
         * The comments.
         */
        comments: ISharedItemComment[]
    }

    /**
     *  Like the OrgRepo type, except with fields that should not be accessible with
     *  a secret URL removed.
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItemOrgRepo {
        __typename: 'SharedItemOrgRepo'

        /**
         * The ID.
         */
        id: number

        /**
         * The remote URI.
         */
        remoteUri: string

        /**
         * See OrgRepo.repository.
         */
        repository: IRepository | null
    }

    /**
     *  Exactly the same as the ThreadLines type, except it cannot have sensitive
     *  fields accidentally added.
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItemThreadLines {
        __typename: 'SharedItemThreadLines'

        /**
         * Returns the HTML before.
         */
        htmlBefore: string

        /**
         * Returns the HTML.
         */
        html: string

        /**
         * Returns the HTML after.
         */
        htmlAfter: string

        /**
         * The text before.
         */
        textBefore: string

        /**
         * The text.
         */
        text: string

        /**
         * The text after.
         */
        textAfter: string

        /**
         * The start of the text selection range.
         */
        textSelectionRangeStart: number

        /**
         * The length of the text selection range.
         */
        textSelectionRangeLength: number
    }

    export interface IHtmlBeforeOnSharedItemThreadLinesArguments {
        isLightTheme: boolean
    }

    export interface IHtmlOnSharedItemThreadLinesArguments {
        isLightTheme: boolean
    }

    export interface IHtmlAfterOnSharedItemThreadLinesArguments {
        isLightTheme: boolean
    }

    /**
     *  Like the Comment type, except with fields that should not be accessible with a
     *  secret URL removed.
     * !
     * ! 🚨 SECURITY: Every field here is accessible publicly given a shared item URL.
     * ! Do NOT use any non-primitive GraphQL type here unless it is also a SharedItem
     * ! type.
     */
    export interface ISharedItemComment {
        __typename: 'SharedItemComment'

        /**
         * The ID.
         */
        id: ID

        /**
         * The database ID.
         */
        databaseID: number

        /**
         * The title.
         */
        title: string

        /**
         * The contents.
         */
        contents: string

        /**
         * The rich HTML.
         */
        richHTML: string

        /**
         * The time when this was created.
         */
        createdAt: string

        /**
         * The time when this was updated.
         */
        updatedAt: string

        /**
         * The author.
         */
        author: ISharedItemUser
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
         * Whether the viewer can reload the site (with the reloadSite mutation).
         */
        canReloadSite: boolean

        /**
         * List all threads. This is an experimental feature.
         */
        threads: IThreadConnection

        /**
         * Lists all language servers.
         */
        langServers: ILangServer[]

        /**
         * The language server for a given language (if exists, otherwise null)
         */
        langServer: ILangServer | null

        /**
         * A list of all access tokens on this site.
         */
        accessTokens: IAccessTokenConnection

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
         * The activity.
         */
        activity: ISiteActivity
    }

    export interface IThreadsOnSiteArguments {
        /**
         * Returns the first n threads from the list.
         */
        first?: number | null
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
         * Messages describing validation errors or usage of deprecated configuration in the configuration JSON
         * (pendingContents if it exists, otherwise effectiveContents). This includes both JSON Schema validation errors
         * and other messages that perform more advanced checks on the configuration (that can't be expressed in the
         * JSON Schema).
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
     * A mutation.
     */
    export interface IMutation {
        __typename: 'Mutation'

        /**
         * Creates a thread. This is an experimental feature.
         */
        createThread: IThread

        /**
         * Creates a thread. This is an experimental feature.
         * @deprecated use createThread instead
         */
        createThread2: IThread

        /**
         *  Updates the user profile information for the user with the given ID.
         *
         *  Only the user and site admins may perform this mutation.
         */
        updateUser: IEmptyResponse

        /**
         *  Updates the user settings for the user with the given ID.
         *
         *  Only the user and site admins may perform this mutation.
         */
        updateUserSettings: ISettings

        /**
         * Update the global settings for all users.
         */
        updateSiteSettings: ISettings

        /**
         * Updates a thread. This is an experimental feature.
         */
        updateThread: IThread

        /**
         * Adds a comment to a thread. This is an experimental feature.
         */
        addCommentToThread: IThread

        /**
         *  This method is the same as addCommentToThread, the only difference is
         *  that authentication is based on the secret ULID instead of the current
         *  user. This is an experimental feature.
         * !
         * ! 🚨 SECURITY: Every field of the return type here is accessible publicly
         * ! given a shared item URL.
         */
        addCommentToThreadShared: ISharedItemThread

        /**
         * Shares a thread. This is an experimental feature.
         */
        shareThread: string

        /**
         * Shares a comment. This is an experimental feature.
         */
        shareComment: string

        /**
         * Creates an org.
         */
        createOrg: IOrg

        /**
         * Updates an org.
         */
        updateOrg: IOrg

        /**
         * Updates an org's settings.
         */
        updateOrgSettings: ISettings

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
         * Creates a user account for a new user and generates a reset password link that the user
         * must visit to sign into the account. Only site admins may perform this mutation.
         */
        createUserBySiteAdmin: ICreateUserBySiteAdminResult

        /**
         * Randomize a user's password so that they need to reset it before they can sign in again.
         * Only site admins may perform this mutation.
         */
        randomizeUserPasswordBySiteAdmin: IRandomizeUserPasswordBySiteAdminResult

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
         *  Invite a user to the organization, either by username or email address.
         *
         *  Only organization members and site admins may perform this mutation.
         */
        inviteUserToOrganization: IInviteUserResult | null

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
         * Accepts a user invite.
         */
        acceptUserInvite: IEmptyResponse | null

        /**
         *  Immediately add a user to the organization, either by username or email address, without sending an
         *  invitation email.
         *
         *  Only site admins may perform this mutation. Organization members may use the inviteUser mutation to invite
         *  users.
         */
        addUserToOrganization: IEmptyResponse

        /**
         * Removes a user from an organization.
         */
        removeUserFromOrg: IEmptyResponse | null

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
    }

    export interface ICreateThreadOnMutationArguments {
        input: ICreateThreadInput
    }

    export interface ICreateThread2OnMutationArguments {
        input: ICreateThreadInput
    }

    export interface IUpdateUserOnMutationArguments {
        user: ID
        username?: string | null
        displayName?: string | null
        avatarURL?: string | null
    }

    export interface IUpdateUserSettingsOnMutationArguments {
        user: ID
        lastKnownSettingsID?: number | null
        contents: string
    }

    export interface IUpdateSiteSettingsOnMutationArguments {
        lastKnownSettingsID?: number | null
        contents: string
    }

    export interface IUpdateThreadOnMutationArguments {
        threadID: ID
        archived?: boolean | null
    }

    export interface IAddCommentToThreadOnMutationArguments {
        threadID: ID
        contents: string
    }

    export interface IAddCommentToThreadSharedOnMutationArguments {
        ulid: string
        threadID: ID
        contents: string
    }

    export interface IShareThreadOnMutationArguments {
        threadID: ID
    }

    export interface IShareCommentOnMutationArguments {
        commentID: ID
    }

    export interface ICreateOrgOnMutationArguments {
        name: string
        displayName: string
    }

    export interface IUpdateOrgOnMutationArguments {
        id: ID
        displayName?: string | null
    }

    export interface IUpdateOrgSettingsOnMutationArguments {
        /**
         * The ID of the org whose settings should be updated.
         */
        id?: ID | null

        /**
         * DEPRECATED: use id instead.
         */
        orgID?: ID | null

        /**
         * The last known settings ID.
         */
        lastKnownSettingsID?: number | null

        /**
         * The contents of the settings.
         */
        contents: string
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

    export interface ICreateUserBySiteAdminOnMutationArguments {
        username: string
        email: string
    }

    export interface IRandomizeUserPasswordBySiteAdminOnMutationArguments {
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

    export interface IInviteUserToOrganizationOnMutationArguments {
        organization: ID
        usernameOrEmail: string
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

    export interface IAcceptUserInviteOnMutationArguments {
        inviteToken: string
    }

    export interface IAddUserToOrganizationOnMutationArguments {
        organization: ID
        usernameOrEmail: string
    }

    export interface IRemoveUserFromOrgOnMutationArguments {
        userID: ID
        orgID: ID
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
     * Input Object for creating threads.
     */
    export interface ICreateThreadInput {
        /**
         * The ID of the organization on which to create the thread.
         */
        orgID: ID

        /**
         * The ID of the canonical remote.
         */
        canonicalRemoteID: string

        /**
         * The clone URL.
         */
        cloneURL: string

        /**
         * The repo revision path.
         */
        repoRevisionPath: string

        /**
         * The lines revision path.
         */
        linesRevisionPath: string

        /**
         * The repo revision.
         */
        repoRevision: string

        /**
         * The lines revision.
         */
        linesRevision: string

        /**
         * The branch.
         */
        branch?: string | null

        /**
         * The start line.
         */
        startLine: number

        /**
         * The end line.
         */
        endLine: number

        /**
         * The start character.
         */
        startCharacter: number

        /**
         * The end character.
         */
        endCharacter: number

        /**
         * The range length.
         */
        rangeLength: number

        /**
         * The contents.
         */
        contents: string

        /**
         * The lines.
         */
        lines?: IThreadLinesInput | null
    }

    /**
     * The Input Object for ThreadLines.
     */
    export interface IThreadLinesInput {
        /**
         * HTML context lines before 'html'.
         */
        htmlBefore: string

        /**
         * HTML lines that the user's selection was made on.
         */
        html: string

        /**
         * HTML context lines after 'html'.
         */
        htmlAfter: string

        /**
         * Text context lines before 'text'.
         */
        textBefore: string

        /**
         * Text lines that the user's selection was made on.
         */
        text: string

        /**
         * Text context lines after 'text'.
         */
        textAfter: string

        /**
         * Byte offset into textLines where user selection began.
         */
        textSelectionRangeStart: number

        /**
         * Length in bytes of the user selection.
         */
        textSelectionRangeLength: number
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
     * The result for Mutation.createUserBySiteAdmin.
     */
    export interface ICreateUserBySiteAdminResult {
        __typename: 'CreateUserBySiteAdminResult'

        /**
         * The reset password URL that the new user must visit to sign into their account.
         */
        resetPasswordURL: string
    }

    /**
     * The result for Mutation.randomizeUserPasswordBySiteAdmin.
     */
    export interface IRandomizeUserPasswordBySiteAdminResult {
        __typename: 'RandomizeUserPasswordBySiteAdminResult'

        /**
         * The reset password URL that the user must visit to sign into their account again.
         */
        resetPasswordURL: string
    }

    /**
     * A invite user result.
     */
    export interface IInviteUserResult {
        __typename: 'InviteUserResult'

        /**
         * The URL that the invited user can visit to accept the invitation.
         */
        acceptInviteURL: string
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
         * The subject whose configuration to mutate (org, user, etc.).
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
         * Perform a raw configuration update. Use one of the other fields on this
         * type instead if possible.
         */
        updateConfiguration: IUpdateConfigurationPayload | null

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

    export interface IUpdateConfigurationOnConfigurationMutationArguments {
        input: IUpdateConfigurationInput
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
     * Input to ConfigurationMutation.updateConfiguration.
     */
    export interface IUpdateConfigurationInput {
        /**
         *  The name of the property to update.
         *
         *  Inserting into an existing array is not yet supported.
         */
        property: string

        /**
         *  The new JSON-encoded value to insert. If the field's value is not set, the property is removed. (This is
         *  different from the field's value being the JSON null value.)
         *
         *  When the value is a non-primitive type, it must be specified using a GraphQL variable, not an inline literal,
         *  or else the GraphQL parser will return an error.
         */
        value?: any | null
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
}
