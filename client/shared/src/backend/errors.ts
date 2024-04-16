import { hasProperty, isErrorLike } from '@sourcegraph/common'

const CLONE_IN_PROGRESS_ERROR_NAME = 'CloneInProgressError'
export class CloneInProgressError extends Error {
    public readonly name = CLONE_IN_PROGRESS_ERROR_NAME
    constructor(repoName: string, public readonly progress?: string) {
        super(`${repoName} is clone in progress`)
    }
}
// Will work even for errors that came from GraphQL, background pages, comlink webworkers, etc.
// TODO remove error message assertion after https://github.com/sourcegraph/sourcegraph/issues/9697 and https://github.com/sourcegraph/sourcegraph/issues/9693 are fixed
export const isCloneInProgressErrorLike = (value: unknown): boolean =>
    isErrorLike(value) && (value.name === CLONE_IN_PROGRESS_ERROR_NAME || /clone in progress/i.test(value.message))

const REPO_NOT_FOUND_ERROR_NAME = 'RepoNotFoundError' as const
export class RepoNotFoundError extends Error {
    public readonly name = REPO_NOT_FOUND_ERROR_NAME
    constructor(repoName: string) {
        super(`repo ${repoName} not found`)
    }
}
// Will work even for errors that came from GraphQL, background pages, comlink webworkers, etc.
// TODO remove error message assertion after https://github.com/sourcegraph/sourcegraph/issues/9697 and https://github.com/sourcegraph/sourcegraph/issues/9693 are fixed
export const isRepoNotFoundErrorLike = (value: unknown): boolean =>
    isErrorLike(value) && (value.name === REPO_NOT_FOUND_ERROR_NAME || /repo.*not found/i.test(value.message))

const REVISION_NOT_FOUND_ERROR_NAME = 'RevisionNotFoundError' as const
export class RevisionNotFoundError extends Error {
    public readonly name = REVISION_NOT_FOUND_ERROR_NAME
    constructor(revision?: string) {
        super(`Revision ${String(revision)} not found`)
    }
}
// Will work even for errors that came from GraphQL, background pages, comlink webworkers, etc.
// TODO remove error message assertion after https://github.com/sourcegraph/sourcegraph/issues/9697 and https://github.com/sourcegraph/sourcegraph/issues/9693 are fixed
export const isRevisionNotFoundErrorLike = (value: unknown): boolean =>
    isErrorLike(value) && (value.name === REVISION_NOT_FOUND_ERROR_NAME || /revision.*not found/i.test(value.message))

const REPO_SEE_OTHER_ERROR_NAME = 'RepoSeeOtherError' as const
export class RepoSeeOtherError extends Error {
    public readonly name = REPO_SEE_OTHER_ERROR_NAME
    constructor(public readonly redirectURL: string) {
        super(`Repository not found at this location, but might exist at ${redirectURL}`)
    }
}

const REPO_DENIED_ERROR_NAME = 'RepoDeniedError' as const
export class RepoDeniedError extends Error {
    public readonly name = REPO_DENIED_ERROR_NAME
    constructor(public readonly reason: string) {
        super(`Repository could not be added on-demand: ${reason}`)
    }
}

export const isRepoDeniedErrorLike = (value: unknown): value is RepoDeniedError =>
    isErrorLike(value) && value.name === REPO_DENIED_ERROR_NAME

// Will work even for errors that came from GraphQL, background pages, comlink webworkers, etc.
// TODO remove error message assertion after https://github.com/sourcegraph/sourcegraph/issues/9697 and https://github.com/sourcegraph/sourcegraph/issues/9693 are fixed
/** Returns the redirect URL if the passed value is like a RepoSeeOtherError, otherwise `false`. */
export const isRepoSeeOtherErrorLike = (value: unknown): string | false => {
    if (!isErrorLike(value)) {
        return false
    }
    if (
        value.name === REPO_SEE_OTHER_ERROR_NAME &&
        hasProperty('redirectURL')(value) &&
        typeof value.redirectURL === 'string'
    ) {
        return value.redirectURL
    }
    const match = value.message.match(/repository not found at this location, but might exist at (\S+)/i)
    if (match) {
        return match[1]
    }
    return false
}
