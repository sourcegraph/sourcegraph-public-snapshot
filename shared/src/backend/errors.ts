export const CLONE_IN_PROGRESS_ERROR_NAME = 'CloneInProgressError'
export class CloneInProgressError extends Error {
    public readonly name = CLONE_IN_PROGRESS_ERROR_NAME
    constructor(repoName: string, public readonly progress?: string) {
        super(`${repoName} is clone in progress`)
    }
}

export const REPO_NOT_FOUND_ERROR_NAME = 'RepoNotFoundError'
export class RepoNotFoundError extends Error {
    public readonly name = REPO_NOT_FOUND_ERROR_NAME
    constructor(repoName: string) {
        super(`repo ${repoName} not found`)
    }
}

export const REV_NOT_FOUND_ERROR_NAME = 'RevNotFoundError'
export class RevNotFoundError extends Error {
    public readonly name = REV_NOT_FOUND_ERROR_NAME
    constructor(rev?: string) {
        super(`rev ${String(rev)} not found`)
    }
}

export const REPO_SEE_OTHER_ERROR_NAME = 'RepoSeeOtherError'
export class RepoSeeOtherError extends Error {
    public readonly name = REPO_SEE_OTHER_ERROR_NAME
    constructor(public readonly redirectURL: string) {
        super(`Repository not found at this location, but might exist at ${redirectURL}`)
    }
}

/**
 * ERPRIVATEREPOPUBLICSOURCEGRAPHCOM means that the current repository is
 * private and the current Sourcegraph URL is Sourcegraph.com. Requests made
 * from a private repository to Sourcegraph.com are blocked unless the
 * `requestMightContainPrivateInfo` argument to `requestGraphQL` is explicitly
 * set to false (defaults to true to be conservative).
 */
export const PRIVATE_REPO_PUBLIC_SOURCEGRAPH_COM_ERROR_NAME = 'PrivateRepoPublicSourcegraphError'
export class PrivateRepoPublicSourcegraphComError extends Error {
    public readonly name = PRIVATE_REPO_PUBLIC_SOURCEGRAPH_COM_ERROR_NAME
    constructor(graphQLName: string) {
        super(
            `A ${graphQLName} GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.`
        )
    }
}

export const AUTH_REQUIRED_ERROR_NAME = 'AuthRequiredError'
