import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    CloneInProgressError,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevNotFoundError,
} from '../../../shared/src/backend/errors'
import { FetchFileCtx } from '../../../shared/src/components/CodeExcerpt'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { createAggregateError } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { AbsoluteRepoFile, makeRepoURI, RepoRev } from '../../../shared/src/util/url'
import { queryGraphQL } from '../backend/graphql'

/**
 * Fetch the repository.
 */
export const fetchRepository = memoizeObservable(
    (args: { repoName: string }): Observable<GQL.IRepository> =>
        queryGraphQL(
            gql`
                query RepositoryRedirect($repoName: String!) {
                    repositoryRedirect(name: $repoName) {
                        __typename
                        ... on Repository {
                            id
                            name
                            url
                            externalURLs {
                                url
                                serviceType
                            }
                            description
                            viewerCanAdminister
                            defaultBranch {
                                displayName
                            }
                        }
                        ... on Redirect {
                            url
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
                if (!data.repositoryRedirect) {
                    throw new RepoNotFoundError(args.repoName)
                }
                if (data.repositoryRedirect.__typename === 'Redirect') {
                    throw new RepoSeeOtherError(data.repositoryRedirect.url)
                }
                return data.repositoryRedirect
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
 *
 * @returns Observable that emits the commit ID. Errors with a `CloneInProgressError` if the repo is still being cloned.
 */
export const resolveRev = memoizeObservable(
    (ctx: { repoName: string; rev?: string }): Observable<ResolvedRev> =>
        queryGraphQL(
            gql`
                query ResolveRev($repoName: String!, $rev: String!) {
                    repositoryRedirect(name: $repoName) {
                        __typename
                        ... on Repository {
                            mirrorInfo {
                                cloneInProgress
                                cloneProgress
                                cloned
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
                        }
                        ... on Redirect {
                            url
                        }
                    }
                }
            `,
            { ...ctx, rev: ctx.rev || '' }
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw createAggregateError(errors)
                }
                if (!data.repositoryRedirect) {
                    throw new RepoNotFoundError(ctx.repoName)
                }
                if (data.repositoryRedirect.__typename === 'Redirect') {
                    throw new RepoSeeOtherError(data.repositoryRedirect.url)
                }
                if (data.repositoryRedirect.mirrorInfo.cloneInProgress) {
                    throw new CloneInProgressError(
                        ctx.repoName,
                        data.repositoryRedirect.mirrorInfo.cloneProgress || undefined
                    )
                }
                if (!data.repositoryRedirect.mirrorInfo.cloned) {
                    throw new CloneInProgressError(ctx.repoName, 'queued for cloning')
                }
                if (!data.repositoryRedirect.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                if (!data.repositoryRedirect.defaultBranch || !data.repositoryRedirect.commit.tree) {
                    throw new RevNotFoundError('HEAD')
                }
                return {
                    commitID: data.repositoryRedirect.commit.oid,
                    defaultBranch: data.repositoryRedirect.defaultBranch.abbrevName,
                    rootTreeURL: data.repositoryRedirect.commit.tree.url,
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
                if (!data?.repository?.commit?.file?.highlight) {
                    throw createAggregateError(errors)
                }
                const file = data.repository.commit.file
                return { isDirectory: file.isDirectory, richHTML: file.richHTML, highlightedFile: file.highlight }
            })
        ),
    ctx => makeRepoURI(ctx) + `?disableTimeout=${String(ctx.disableTimeout)}&isLightTheme=${String(ctx.isLightTheme)}`
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
                let parsed = result.highlightedFile.html.substr('<table>'.length)
                parsed = parsed.substr(0, parsed.length - '</table>'.length)
                const rows = parsed.split('</tr>')
                for (let i = 0; i < rows.length; ++i) {
                    rows[i] += '</tr>'
                }
                return rows
            })
        ),
    ctx => makeRepoURI(ctx) + `?isLightTheme=${String(ctx.isLightTheme)}`
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
                if (!data?.repository?.commit?.file?.externalURLs) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.file.externalURLs
            })
        ),
    makeRepoURI
)

export const fetchTreeEntries = memoizeObservable(
    (args: AbsoluteRepoFile & { first?: number }): Observable<GQL.IGitTree> =>
        queryGraphQL(
            gql`
                query TreeEntries(
                    $repoName: String!
                    $rev: String!
                    $commitID: String!
                    $filePath: String!
                    $first: Int
                ) {
                    repository(name: $repoName) {
                        commit(rev: $commitID, inputRevspec: $rev) {
                            tree(path: $filePath) {
                                isRoot
                                url
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
                if (errors || !data?.repository?.commit?.tree) {
                    throw createAggregateError(errors)
                }
                return data.repository.commit.tree
            })
        ),
    ({ first, ...args }) => `${makeRepoURI(args)}:first-${String(first)}`
)
