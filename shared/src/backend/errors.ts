import { asError, ErrorLike } from '../util/errors'

export const EAGGREGATE = 'AggregateError'
/**
 * An Error that aggregates multiple errors
 */
export class AggregateError extends Error {
    public readonly name = EAGGREGATE
    public readonly code = EAGGREGATE
    constructor(public readonly errors: ErrorLike[] = []) {
        super(errors.map(({ message }) => message).join('\n'))
        this.errors = errors.map(asError)
    }
}

export const ECLONEINPROGESS = 'CloneInProgressError'
export class CloneInProgressError extends Error {
    public readonly name = ECLONEINPROGESS
    public readonly code = ECLONEINPROGESS
    constructor(repoName: string, public readonly progress?: string) {
        super(`${repoName} is clone in progress`)
    }
}

export const EREPONOTFOUND = 'RepoNotFoundError'
export class RepoNotFoundError extends Error {
    public readonly name = EREPONOTFOUND
    public readonly code = EREPONOTFOUND
    constructor(repoName: string) {
        super(`repo ${repoName} not found`)
    }
}

export const EREVNOTFOUND = 'RevNotFoundError'
export class RevNotFoundError extends Error {
    public readonly name = EREVNOTFOUND
    public readonly code = EREVNOTFOUND
    constructor(rev?: string) {
        super(`rev ${String(rev)} not found`)
    }
}

export const EREPOSEEOTHER = 'ERREPOSEEOTHER'
export class RepoSeeOtherError extends Error {
    public readonly name = EREPOSEEOTHER
    public readonly code = EREPOSEEOTHER
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
export const ERPRIVATEREPOPUBLICSOURCEGRAPHCOM = 'PrivateRepoPublicSourcegraph'
export class PrivateRepoPublicSourcegraphComError extends Error {
    public readonly name = ERPRIVATEREPOPUBLICSOURCEGRAPHCOM
    public readonly code = ERPRIVATEREPOPUBLICSOURCEGRAPHCOM
    constructor(graphQLName: string) {
        super(
            `A ${graphQLName} GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.`
        )
    }
}

export const ERAUTHREQUIRED = 'AuthRequiredError'
