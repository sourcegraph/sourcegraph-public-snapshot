import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { gql, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { createAggregateError } from '../util/errors'
import { memoizeObservable } from '../util/memoize'
import { makeRepoURI } from './index'

// We don't subclass Error because Error is not subclassable in ES5.
// Use the internal factory functions and check for the error code on callsites.

export const ECLONEINPROGESS = 'ECLONEINPROGESS'
export interface CloneInProgressError extends Error {
    code: typeof ECLONEINPROGESS
    progress?: string
}
const createCloneInProgressError = (repoPath: string, progress: string | undefined): CloneInProgressError =>
    Object.assign(new Error(`Repository ${repoPath} is clone in progress`), {
        code: ECLONEINPROGESS as typeof ECLONEINPROGESS,
        progress,
    })

export const EREPONOTFOUND = 'EREPONOTFOUND'
const createRepoNotFoundError = (repoPath: string): Error =>
    Object.assign(new Error(`Repository ${repoPath} not found`), { code: EREPONOTFOUND })

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
    (args: { repoPath: string }): Observable<GQL.IRepository> =>
        queryGraphQL(
            gql`
                query Repository($repoPath: String!) {
                    repository(uri: $repoPath) {
                        id
                        uri
                        externalURLs {
                            url
                            serviceType
                        }
                        description
                        enabled
                        viewerCanAdminister
                        redirectURL
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
                    throw createRepoNotFoundError(args.repoPath)
                }
                return data.repository
            })
        ),
    makeRepoURI
)

export interface ResolvedRev {
    commitID: string
    defaultBranch: string
}

/**
 * When `rev` is undefined, the default branch is resolved.
 * @return Observable that emits the commit ID
 *         Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    (ctx: { repoPath: string; rev?: string }): Observable<ResolvedRev> =>
        queryGraphQL(
            gql`
                query ResolveRev($repoPath: String!, $rev: String!) {
                    repository(uri: $repoPath) {
                        mirrorInfo {
                            cloneInProgress
                            cloneProgress
                        }
                        commit(rev: $rev) {
                            oid
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
                    throw createRepoNotFoundError(ctx.repoPath)
                }
                if (data.repository.mirrorInfo.cloneInProgress) {
                    throw createCloneInProgressError(
                        ctx.repoPath,
                        data.repository.mirrorInfo.cloneProgress || undefined
                    )
                }
                if (!data.repository.commit) {
                    throw createRevNotFoundError(ctx.rev)
                }
                if (!data.repository.defaultBranch) {
                    throw createRevNotFoundError('HEAD')
                }
                return {
                    commitID: data.repository.commit.oid,
                    defaultBranch: data.repository.defaultBranch.abbrevName,
                }
            })
        ),
    makeRepoURI
)

interface FetchFileCtx {
    repoPath: string
    commitID: string
    filePath: string
    disableTimeout?: boolean
    isLightTheme: boolean
}

interface HighlightedFileResult {
    isDirectory: boolean
    richHTML: string
    highlightedFile: GQL.IHighlightedFile
}

export const fetchHighlightedFile = memoizeObservable(
    (ctx: FetchFileCtx): Observable<HighlightedFileResult> =>
        queryGraphQL(
            gql`
                query HighlightedFile(
                    $repoPath: String!
                    $commitID: String!
                    $filePath: String!
                    $disableTimeout: Boolean!
                    $isLightTheme: Boolean!
                ) {
                    repository(uri: $repoPath) {
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

interface BlobContent {
    isDirectory: boolean
    content: string
}

export const fetchBlobContent = memoizeObservable(
    (ctx: FetchFileCtx): Observable<BlobContent> =>
        queryGraphQL(
            gql`
                query BlobContent($repoPath: String!, $commitID: String!, $filePath: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                isDirectory
                                content
                            }
                        }
                    }
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.repository || !data.repository.commit || !data.repository.commit.file) {
                    throw createAggregateError(errors)
                }
                const file = data.repository.commit.file
                return { isDirectory: file.isDirectory, content: file.content }
            })
        ),
    makeRepoURI
)

interface FetchFileMetadataCtx {
    repoPath: string
    rev: string
    filePath: string
}

export const fetchFileExternalLinks = memoizeObservable(
    (ctx: FetchFileMetadataCtx): Observable<GQL.IExternalLink[]> =>
        queryGraphQL(
            gql`
                query FileExternalLinks($repoPath: String!, $rev: String!, $filePath: String!) {
                    repository(uri: $repoPath) {
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
