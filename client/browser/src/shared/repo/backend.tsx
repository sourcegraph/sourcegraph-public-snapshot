import { Observable } from 'rxjs'
import { catchError, delay, filter, map, retryWhen } from 'rxjs/operators'
import { AbsoluteRepoFile, makeRepoURI } from '.'
import { getContext } from '../backend/context'
import { CloneInProgressError, ECLONEINPROGESS, RepoNotFoundError, RevNotFoundError } from '../backend/errors'
import { queryGraphQL } from '../backend/graphql'
import { memoizeObservable } from '../util/memoize'

/**
 * @return Observable that emits the repo URL
 *         Errors with a `RepoNotFoundError` if the repo is not found
 */
export const resolveRepo = memoizeObservable(
    (ctx: { repoPath: string }): Observable<string> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query ResolveRepo($repoPath: String!) {
                repository(name: $repoPath) {
                    url
                }
            }`,
            variables: { ...ctx },
        }).pipe(
            map(result => {
                if (!result.data || !result.data.repository) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }

                return result.data.repository.url
            }, catchError((err, caught) => caught))
        ),
    makeRepoURI
)

/**
 * @return Observable that emits the commit ID
 *         Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    (ctx: { repoPath: string; rev?: string }): Observable<string> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query ResolveRev($repoPath: String!, $rev: String!) {
                repository(name: $repoPath) {
                    mirrorInfo {
                        cloneInProgress
                    }
                    commit(rev: $rev) {
                        oid
                    }
                }
            }`,
            variables: { ...ctx, rev: ctx.rev || '' },
        }).pipe(
            map(result => {
                if (!result.data) {
                    throw new Error('invalid response received from graphql endpoint')
                }
                if (!result.data.repository) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }
                if (result.data.repository.mirrorInfo.cloneInProgress) {
                    throw new CloneInProgressError(ctx.repoPath)
                }
                if (!result.data.repository.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                return result.data.repository.commit.oid
            }),
            catchError(error => {
                if (
                    error &&
                    error.data &&
                    error.data.repository &&
                    error.data.repository.mirrorInfo &&
                    error.data.repository.mirrorInfo.cloneInProgress
                ) {
                    throw new CloneInProgressError(ctx.repoPath)
                }

                throw error
            })
        ),
    makeRepoURI
)

export function retryWhenCloneInProgressError<T>(): (v: Observable<T>) => Observable<T> {
    return (maybeErrors: Observable<T>) =>
        maybeErrors.pipe(
            retryWhen(errors =>
                errors.pipe(
                    filter(err => {
                        if (err.code === ECLONEINPROGESS) {
                            return true
                        }

                        // Don't swollow other errors.
                        throw err
                    }),
                    delay(1000)
                )
            )
        )
}

const trimRepoPath = ({ repoPath, ...rest }) => ({ ...rest, repoPath: repoPath.replace(/.git$/, '') })

export const fetchBlobContentLines = memoizeObservable(
    (ctx: AbsoluteRepoFile): Observable<string[]> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query BlobContent($repoPath: String!, $commitID: String!, $filePath: String!) {
                repository(name: $repoPath) {
                    commit(rev: $commitID) {
                        file(path: $filePath) {
                            content
                        }
                    }
                }
            }`,
            variables: trimRepoPath(ctx),
            retry: false,
        }).pipe(
            map(({ data }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.file ||
                    !data.repository.commit.file.content
                ) {
                    return []
                }
                return data.repository.commit.file!.content.split('\n')
            }),
            catchError(({ errors, ...rest }) => {
                if (errors && errors.length === 1) {
                    const err = errors[0]
                    const isFileContent = err.path.join('.') === 'repository.commit.file.content'
                    const isDNE = /does not exist/.test(err.message)

                    // The error is the file DNE. Just ignore it and pass an empty array
                    // to represent this.
                    if (isFileContent && isDNE) {
                        return []
                    }
                }

                // Don't swollow unexpected errors
                throw { errors, ...rest }
            })
        ),
    makeRepoURI
)
