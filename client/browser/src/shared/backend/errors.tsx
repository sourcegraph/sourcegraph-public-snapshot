export interface ErrorLike {
    message?: string
    code?: string
}

export const isErrorLike = (val: any): val is ErrorLike =>
    !!val && typeof val === 'object' && (!!val.stack || ('message' in val || 'code' in val))

/**
 * Converts an ErrorLike to a proper Error if needed, copying all properties
 * @param errorLike An Error or object with ErrorLike properties
 */
export const asError = (errorLike: ErrorLike): Error =>
    errorLike instanceof Error ? errorLike : Object.assign(new Error(errorLike.message), errorLike)

/**
 * An Error that aggregates multiple errors
 */
export interface AggregateError extends Error {
    name: 'AggregateError'
    errors: Error[]
}

/**
 * Creates an aggregate error out of multiple provided error likes
 *
 * @param errors The errors or ErrorLikes to aggregate
 */
export const createAggregateError = (errors: ErrorLike[] = []): AggregateError =>
    Object.assign(new Error(errors.map(e => e.message).join('\n')), {
        name: 'AggregateError' as 'AggregateError',
        errors: errors.map(asError),
    })

/**
 * Improves error messages in case of ajax errors
 */
export const normalizeAjaxError = (err: any): void => {
    if (!err) {
        return
    }
    if (typeof err.status === 'number') {
        if (err.status === 0) {
            err.message = 'Unable to reach server. Check your network connection and try again in a moment.'
        } else {
            err.message = `Unexpected HTTP error: ${err.status}`
            if (err.xhr && err.xhr.statusText) {
                err.message += ` ${err.xhr.statusText}`
            }
        }
    }
}

export const ECLONEINPROGESS = 'ECLONEINPROGESS'
export class CloneInProgressError extends Error {
    public readonly code = ECLONEINPROGESS
    constructor(repoPath: string) {
        super(`${repoPath} is clone in progress`)
    }
}

export const EREPONOTFOUND = 'EREPONOTFOUND'
export class RepoNotFoundError extends Error {
    public readonly code = EREPONOTFOUND
    constructor(repoPath: string) {
        super(`repo ${repoPath} not found`)
    }
}

export const EREVNOTFOUND = 'EREVNOTFOUND'
export class RevNotFoundError extends Error {
    public readonly code = EREVNOTFOUND
    constructor(rev?: string) {
        super(`rev ${rev} not found`)
    }
}

export const ERNOSOURCEGRAPHURL = 'ERNOSOURCEGRAPHURL'
export class NoSourcegraphURLError extends Error {
    public readonly code = ERNOSOURCEGRAPHURL
    constructor() {
        super(`no sourcegraph urls are configured`)
    }
}

/**
 * ERPRIVATEREPOPUBLICSOURCEGRAPHCOM means that the current repository is
 * private and the current Sourcegraph URL is Sourcegraph.com. Requests made
 * from a private repository to Sourcegraph.com are blocked unless the
 * `requestMightContainPrivateInfo` argument to `requestGraphQL` is explicitly
 * set to false (defaults to true to be conservative).
 */
export const ERPRIVATEREPOPUBLICSOURCEGRAPHCOM = 'ERPRIVATEREPOPUBLICSOURCEGRAPHCOM'
export class PrivateRepoPublicSourcegraphComError extends Error {
    public readonly code = ERPRIVATEREPOPUBLICSOURCEGRAPHCOM
    constructor(graphQLName: string) {
        super(
            `A ${graphQLName} GraphQL request to the public Sourcegraph.com was blocked because the current repository is private.`
        )
    }
}

export const ERAUTHREQUIRED = 'ERAUTHREQUIRED'
export interface AuthRequiredError extends Error {
    code: typeof ERAUTHREQUIRED
    url: string
}

export const createAuthRequiredError = (url: string): AuthRequiredError =>
    Object.assign(new Error(`private mode requires authentication: ${url}`), {
        code: ERAUTHREQUIRED as typeof ERAUTHREQUIRED,
        url,
    })
