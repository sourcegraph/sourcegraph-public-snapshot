import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { FetchFileCtx } from '../../../shared/src/components/CodeExcerpt'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { AbsoluteRepoFile, makeRepoURI, RepoRev } from '../../../shared/src/util/url'
import { queryGraphQL } from '../backend/graphql'

// We don't subclass Error because Error is not subclassable in ES5.
// Use the internal factory functions and check for the error code on callsites.

export const ECLONEINPROGESS = 'ECLONEINPROGESS'
export interface CloneInProgressError extends Error {
    code: typeof ECLONEINPROGESS
    progress?: string
}
const createCloneInProgressError = (repoName: string, progress: string | undefined): CloneInProgressError =>
    Object.assign(new Error(`Repository ${repoName} is clone in progress`), {
        code: ECLONEINPROGESS as typeof ECLONEINPROGESS,
        progress,
    })

export const EREPONOTFOUND = 'EREPONOTFOUND'
const createRepoNotFoundError = (repoName: string): Error =>
    Object.assign(new Error(`Repository ${repoName} not found`), { code: EREPONOTFOUND })

export const EREVNOTFOUND = 'EREVNOTFOUND'
const createRevNotFoundError = (rev?: string): Error =>
    Object.assign(new Error(`Revision ${rev} not found`), { code: EREVNOTFOUND })

export const EREPOSEEOTHER = 'ERREPOSEEOTHER'
export interface RepoSeeOtherError extends Error {
    code: typeof EREPOSEEOTHER
    redirectURL: string
}
const createRepoSeeOtherError = (redirectURL: string): RepoSeeOtherError =>
    Object.assign(new Error(`Repository not found at this location, but might exist at ${redirectURL}`), {
        code: EREPOSEEOTHER as typeof EREPOSEEOTHER,
        redirectURL,
    })

/**
 * Fetch the repository.
 */
export const fetchRepository = memoizeObservable(
    (args: { repoName: string }): Observable<GQL.IRepository> =>
        queryGraphQL(
            gql`
                query Repository($repoName: String!) {
                    repository(name: $repoName) {
                        id
                        name
                        url
                        externalURLs {
                            url
                            serviceType
                        }
                        description
                        enabled
                        viewerCanAdminister
                        redirectURL
                        defaultBranch {
                            displayName
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw createAggregateError(errors)
                }
                if (data.repository && data.repository.redirectURL) {
                    throw createRepoSeeOtherError(data.repository.redirectURL)
                }
                if (!data.repository) {
                    throw createRepoNotFoundError(args.repoName)
                }
                return data.repository
            })
        ),
    makeRepoURI
)

export interface ResolvedRev {
    commitID: string
    defaultBranch: string

    /** The URL to the repository root tree at the revision. */
    rootTreeURL: string
}

/**
 * When `rev` is undefined, the default branch is resolved.
 * @return Observable that emits the commit ID
 *         Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    (ctx: { repoName: string; rev?: string }): Observable<ResolvedRev> =>
        queryGraphQL(
            gql`
                query ResolveRev($repoName: String!, $rev: String!) {
                    repository(name: $repoName) {
                        mirrorInfo {
                            cloneInProgress
                            cloneProgress
                        }
                        commit(rev: $rev) {
                            oid
                            tree(path: "") {
                                url
                            }
                        }
                        defaultBranch {
                            abbrevName
                        }
                        redirectURL
                    }
                }
            `,
            { ...ctx, rev: ctx.rev || '' }
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw createAggregateError(errors)
                }
                if (data.repository && data.repository.redirectURL) {
                    throw createRepoSeeOtherError(data.repository.redirectURL)
                }
                if (!data.repository) {
                    throw createRepoNotFoundError(ctx.repoName)
                }
                if (data.repository.mirrorInfo.cloneInProgress) {
                    throw createCloneInProgressError(
                        ctx.repoName,
                        data.repository.mirrorInfo.cloneProgress || undefined
                    )
                }
                if (!data.repository.commit) {
                    throw createRevNotFoundError(ctx.rev)
                }
                if (!data.repository.defaultBranch || !data.repository.commit.tree) {
                    throw createRevNotFoundError('HEAD')
                }
                return {
                    commitID: data.repository.commit.oid,
                    defaultBranch: data.repository.defaultBranch.abbrevName,
                    rootTreeURL: data.repository.commit.tree.url,
                }
            })
        ),
    makeRepoURI
)

interface HighlightedFileResult {
    isDirectory: boolean
    richHTML: string
    highlightedFile: GQL.IHighlightedFile
}

const fetchHighlightedFile = memoizeObservable(
    (ctx: FetchFileCtx): Observable<HighlightedFileResult> =>
        queryGraphQL(
            gql`
                query HighlightedFile(
                    $repoName: String!
                    $commitID: String!
                    $filePath: String!
                    $disableTimeout: Boolean!
                    $isLightTheme: Boolean!
                ) {
                    repository(name: $repoName) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                isDirectory
                                richHTML
                                highlight(disableTimeout: $disableTimeout, isLightTheme: $isLightTheme) {
                                    aborted
                                    html
                                }
                            }
                        }
                    }
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.file ||
                    !data.repository.commit.file.highlight
                ) {
                    throw createAggregateError(errors)
                }
                const file = data.repository.commit.file
                return { isDirectory: file.isDirectory, richHTML: file.richHTML, highlightedFile: file.highlight }
            })
        ),
    ctx => makeRepoURI(ctx) + `?disableTimeout=${ctx.disableTimeout} ` + `?isLightTheme=${ctx.isLightTheme}`
)

/**
 * Produces a list like ['<tr>...</tr>', ...]
 */
export const fetchHighlightedFileLines = memoizeObservable(
    (ctx: FetchFileCtx, force?: boolean): Observable<string[]> =>
        fetchHighlightedFile(ctx, force).pipe(
            map(result => {
                if (result.isDirectory) {
                    return []
                }
                if (result.highlightedFile.aborted) {
                    throw new Error('aborted fetching highlighted contents')
                }
                let parsed = result.highlightedFile.html.substr('<table>'.length)
                parsed = parsed.substr(0, parsed.length - '</table>'.length)
                const rows = parsed.split('</tr>')
                for (let i = 0; i < rows.length; ++i) {
                    rows[i] += '</tr>'
                }
                return rows
            })
        ),
    ctx => makeRepoURI(ctx) + `?isLightTheme=${ctx.isLightTheme}`
)

export const fetchFileExternalLinks = memoizeObservable(
    (ctx: RepoRev & { filePath: string }): Observable<GQL.IExternalLink[]> =>
        queryGraphQL(
            gql`
                query FileExternalLinks($repoName: String!, $rev: String!, $filePath: String!) {
                    repository(name: $repoName) {
                        commit(rev: $rev) {
                            file(path: $filePath) {
                                externalURLs {
                                    url
                                    serviceType
                                }
                            }
                        }
                    }
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.file ||
                    !data.repository.commit.file.externalURLs
                ) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.file.externalURLs
            })
        ),
    makeRepoURI
)

export const fetchTree = memoizeObservable(
    (args: AbsoluteRepoFile & { first?: number }): Observable<GQL.IGitTree> =>
        queryGraphQL(
            gql`
                query Tree($repoName: String!, $rev: String!, $commitID: String!, $filePath: String!, $first: Int) {
                    repository(name: $repoName) {
                        commit(rev: $commitID, inputRevspec: $rev) {
                            tree(path: $filePath) {
                                isRoot
                                url
                                entries(first: $first) {
                                    name
                                    path
                                    isDirectory
                                    url
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || errors || !data.repository || !data.repository.commit || !data.repository.commit.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree
            })
        ),
    makeRepoURI
)

export const fetchTreeEntries = memoizeObservable(
    (args: AbsoluteRepoFile & { first?: number }): Observable<GQL.IGitTree> =>
        queryGraphQL(
            gql`
                query Tree($repoName: String!, $rev: String!, $commitID: String!, $filePath: String!, $first: Int) {
                    repository(name: $repoName) {
                        commit(rev: $commitID, inputRevspec: $rev) {
                            tree(path: $filePath) {
                                entries(first: $first, recursiveSingleChild: true) {
                                    name
                                    path
                                    isDirectory
                                    url
                                    submodule {
                                        url
                                        commit
                                    }
                                    isSingleChild
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data || errors || !data.repository || !data.repository.commit || !data.repository.commit.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree
            })
        ),
    makeRepoURI
)
