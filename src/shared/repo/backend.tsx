import { Observable } from 'rxjs'
import { catchError, delay, filter, map, retryWhen } from 'rxjs/operators'
import { AbsoluteRepoFile, makeRepoURI, parseBrowserRepoURL } from '.'
import { GQL } from '../../types/gqlschema'
import { getContext } from '../backend/context'
import {
    CloneInProgressError,
    createAggregateError,
    ECLONEINPROGESS,
    RepoNotFoundError,
    RevNotFoundError,
} from '../backend/errors'
import { queryGraphQL } from '../backend/graphql'
import { memoizeAsync, memoizeObservable } from '../util/memoize'

/**
 * Fetches the language server for a given language
 *
 * @return Observable that emits the language server for the given language or null if not exists
 */
export function fetchLangServer(
    language: string
): Observable<Pick<GQL.ILangServer, 'displayName' | 'homepageURL' | 'issuesURL' | 'experimental'> | null> {
    return queryGraphQL({
        ctx: getContext({}),
        request: `
            query LangServer($language: String!) {
                site {
                    langServer(language: $language) {
                        displayName
                        homepageURL
                        issuesURL
                        experimental
                    }
                }
            }
        `,
        variables: { language },
    }).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw createAggregateError(errors)
            }
            return data.site.langServer
        })
    )
}

/**
 * @return Observable that emits the parent commit ID for a given commit ID.
 *         Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveParentRev = memoizeObservable(
    (ctx: { repoPath: string; rev: string }): Observable<string> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query ResolveParentRev($repoPath: String!, $rev: String!) {
                repository(uri: $repoPath) {
                    mirrorInfo {
                        cloneInProgress
                    }
                    commit(rev: $rev) {
                        parents {
                            oid
                        }
                    }
                }
            }`,
            variables: { ...ctx, rev: ctx.rev || '' },
        }).pipe(
            map(result => {
                if (!result.data) {
                    throw new Error('invalid response received from graphql endpoint')
                }
                if (!result.data.repository || !result.data.repository.commit) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }
                if (result.data.repository.mirrorInfo.cloneInProgress) {
                    throw new CloneInProgressError(ctx.repoPath)
                }
                if (!result.data.repository.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                if (!result.data.repository.commit.parents) {
                    throw new RevNotFoundError(ctx.rev)
                }
                return result.data.repository.commit.parents[0].oid
            })
        ),
    makeRepoURI
)

/**
 * @return Observable that emits the repo URL
 *         Errors with a `RepoNotFoundError` if the repo is not found
 */
export const resolveRepo = memoizeObservable(
    (ctx: { repoPath: string }): Observable<string> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query ResolveRepo($repoPath: String!) {
                repository(uri: $repoPath) {
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
                repository(uri: $repoPath) {
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

export const listAllSearchResults = memoizeAsync(
    (ctx: { query: string }): Promise<number> =>
        queryGraphQL({
            ctx: getContext({ repoKey: '' }),
            request: `query Search($query: String!) {
                search(query: $query) {
                    results {
                        resultCount
                    }
                }
            }`,
            variables: ctx,
        })
            .toPromise()
            .then(result => {
                if (!result.data || !result.data.search || !result.data.search.results) {
                    throw new Error('invalid response received from graphql endpoint')
                }
                return result.data.search.results.resultCount
            }),
    () => {
        const { repoPath } = parseBrowserRepoURL(window.location.href, window)
        return `${repoPath}:${window.location.search}`
    }
)

const trimRepoPath = ({ repoPath, ...rest }) => ({ ...rest, repoPath: repoPath.replace(/.git$/, '') })

export const fetchBlobContentLines = memoizeObservable(
    (ctx: AbsoluteRepoFile): Observable<string[]> =>
        queryGraphQL({
            ctx: getContext({ repoKey: ctx.repoPath }),
            request: `query BlobContent($repoPath: String!, $commitID: String!, $filePath: String!) {
                repository(uri: $repoPath) {
                    commit(rev: $commitID) {
                        file(path: $filePath) {
                            content
                        }
                    }
                }
            }`,
            variables: trimRepoPath(ctx),
        }).pipe(
            map(({ data, errors }) => {
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
