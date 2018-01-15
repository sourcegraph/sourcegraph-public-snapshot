import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, queryGraphQL } from '../backend/graphql'
import { memoizeObservable } from '../util/memoize'
import { makeRepoURI } from './index'

export const ECLONEINPROGESS = 'ECLONEINPROGESS'
class CloneInProgressError extends Error {
    public readonly code = ECLONEINPROGESS
    constructor(repoPath: string) {
        super(`${repoPath} is clone in progress`)
    }
}

export const EREPONOTFOUND = 'EREPONOTFOUND'
class RepoNotFoundError extends Error {
    public readonly code = EREPONOTFOUND
    constructor(repoPath: string) {
        super(`repo ${repoPath} not found`)
    }
}

export const EREVNOTFOUND = 'EREVNOTFOUND'
class RevNotFoundError extends Error {
    public readonly code = EREVNOTFOUND
    constructor(rev?: string) {
        super(`rev ${rev} not found`)
    }
}

export const ERREPOSEEOTHER = 'ERREPOSEEOTHER'
export interface RepoSeeOtherError {
    redirectURL: string
}
class RepoSeeOtherErrorImpl extends Error implements RepoSeeOtherErrorImpl {
    public readonly code = ERREPOSEEOTHER
    constructor(public redirectURL: string) {
        super(`repo not found at this location, but might exist at ${redirectURL}`)
    }
}

/**
 * Fetch the repository.
 */
export const fetchRepository = memoizeObservable(
    (args: { repoPath: string }): Observable<GQL.IRepository | null> =>
        queryGraphQL(
            gql`
                query Repository($repoPath: String!) {
                    repository(uri: $repoPath) {
                        id
                        uri
                        description
                        viewerCanAdminister
                        redirectURL
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                if (data.repository && data.repository.redirectURL) {
                    throw new RepoSeeOtherErrorImpl(data.repository.redirectURL)
                }
                if (!data.repository) {
                    throw new RepoNotFoundError(args.repoPath)
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
                query ResolveRev($repoPath: String, $rev: String) {
                    repository(uri: $repoPath) {
                        cloneInProgress
                        commit(rev: $rev) {
                            oid
                        }
                        defaultBranch
                        redirectURL
                    }
                }
            `,
            { ...ctx, rev: ctx.rev || '' }
        ).pipe(
            map(({ data, errors }) => {
                if (!data) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                if (data.repository && data.repository.redirectURL) {
                    throw new RepoSeeOtherErrorImpl(data.repository.redirectURL)
                }
                if (!data.repository) {
                    throw new RepoNotFoundError(ctx.repoPath)
                }
                if (data.repository.cloneInProgress) {
                    throw new CloneInProgressError(ctx.repoPath)
                }
                if (!data.repository.commit) {
                    throw new RevNotFoundError(ctx.rev)
                }
                if (!data.repository.defaultBranch) {
                    throw new RevNotFoundError('HEAD')
                }
                return {
                    commitID: data.repository.commit.oid,
                    defaultBranch: data.repository.defaultBranch,
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
                    $repoPath: String
                    $commitID: String
                    $filePath: String
                    $disableTimeout: Boolean
                    $isLightTheme: Boolean
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
                    throw Object.assign(
                        new Error('Could not fetch highlighted file: ' + (errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
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

export const listAllFiles = memoizeObservable(
    (ctx: { repoPath: string; commitID: string }): Observable<string[]> =>
        queryGraphQL(
            gql`
                query FileTree($repoPath: String!, $commitID: String!) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            tree(recursive: true) {
                                files {
                                    name
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
                    !data.repository.commit.tree ||
                    !data.repository.commit.tree.files
                ) {
                    throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
                }
                return data.repository.commit.tree.files.map(file => file.name)
            })
        ),
    makeRepoURI
)

interface BlobContent {
    isDirectory: boolean
    content: string
}

export const fetchBlobContent = memoizeObservable(
    (ctx: FetchFileCtx): Observable<BlobContent> =>
        queryGraphQL(
            gql`
                query BlobContent($repoPath: String, $commitID: String, $filePath: String) {
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
                    throw Object.assign(
                        'Could not fetch blob content: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                const file = data.repository.commit.file
                return { isDirectory: file.isDirectory, content: file.content }
            })
        ),
    makeRepoURI
)

export const fetchPhabricatorRepo = memoizeObservable(
    (ctx: { repoPath: string }): Observable<GQL.IPhabricatorRepo | null> =>
        queryGraphQL(
            gql`
                query PhabricatorRepo($repoPath: String) {
                    phabricatorRepo(uri: $repoPath) {
                        callsign
                        uri
                        url
                    }
                }
            `,
            ctx
        ).pipe(
            map(result => {
                if (result.errors || !result.data || !result.data.phabricatorRepo) {
                    return null
                }
                return result.data.phabricatorRepo
            })
        ),
    makeRepoURI
)
