import { redirect, error } from '@sveltejs/kit'

import { asError, loadMarkdownSyntaxHighlighting, type ErrorLike } from '$lib/common'
import { getGraphQLClient, type GraphQLClient } from '$lib/graphql'
import {
    CloneInProgressError,
    RepoNotFoundError,
    RepoSeeOtherError,
    RevisionNotFoundError,
    displayRepoName,
    isRepoSeeOtherErrorLike,
    isRevisionNotFoundErrorLike,
    parseRepoRevision,
} from '$lib/shared'

import type { LayoutLoad } from './$types'
import { ResolveRepoRevision, ResolvedRepository, type ResolveRepoRevisionResult } from './layout.gql'

export interface ResolvedRevision {
    repo: ResolvedRepository & NonNullable<{ commit: ResolvedRepository['commit'] }>
    commitID: string
    defaultBranch: string
}

export const load: LayoutLoad = async ({ params, url, depends }) => {
    const client = getGraphQLClient()

    // This allows other places to reload all repo related data by calling
    // invalidate('repo:root')
    depends('repo:root')

    // Repo pages render markdown, so we ensure that syntax highlighting for code blocks
    // inside markdown is loaded.
    loadMarkdownSyntaxHighlighting()

    // An empty revision means we are at the default branch
    const { repoName, revision = '' } = parseRepoRevision(params.repo)

    let resolvedRevisionOrError: ResolvedRevision | ErrorLike
    let resolvedRevision: ResolvedRevision | undefined

    try {
        resolvedRevisionOrError = await resolveRepoRevision({ client, repoName, revspec: revision })
        resolvedRevision = resolvedRevisionOrError
    } catch (repoError: unknown) {
        const redirect = isRepoSeeOtherErrorLike(repoError)

        if (redirect) {
            redirectToExternalHost(redirect, url)
        }

        // TODO: use different error codes for different types of errors
        // Let revision errors be handled by the nested layout so that we can
        // still render the main repo navigation and header
        if (!isRevisionNotFoundErrorLike(repoError)) {
            error(400, asError(repoError))
        }

        resolvedRevisionOrError = asError(repoError)
    }

    return {
        repoURL: '/' + params.repo,
        repoURLWithoutRevision: '/' + repoName,
        repoName,
        displayRepoName: displayRepoName(repoName),
        /**
         * Revision from URL
         */
        revision,
        /**
         * A friendly display form of the revision. This can be:
         * - an empty string which signifies the default branch
         * - an abbreviated commit SHA
         * - a symbolic revision (e.g. a branch or tag name)
         */
        displayRevision: displayRevision(revision, resolvedRevision),
        resolvedRevisionOrError,
        resolvedRevision,
    }
}

/**
 * Returns a friendly display form of the revision. If the resolved revision's commit ID starts with the revision,
 * the first 7 characters of the commit ID are returned. Otherwise, the revision is returned as is.
 *
 * @param revision The revision from the URL
 * @param resolvedRevision The resolved revision
 * @returns A human readable revision string
 */
function displayRevision(revision: string, resolvedRevision: ResolvedRevision | undefined): string {
    if (!resolvedRevision) {
        return revision
    }

    if (revision && resolvedRevision.commitID.startsWith(revision)) {
        return resolvedRevision.commitID.slice(0, 7)
    }

    return revision
}

function redirectToExternalHost(externalRedirectURL: string, currentURL: URL): never {
    const externalHostURL = new URL(externalRedirectURL)
    const redirectURL = new URL(currentURL)
    // Preserve the path of the current URL and redirect to the repo on the external host.
    redirectURL.host = externalHostURL.host
    redirectURL.protocol = externalHostURL.protocol
    redirect(303, redirectURL.toString())
}

// This is a cache for resolved repository information to help in the following case:
// - The user navigates to a repository page with a symbolic revspec (e.g. a branch or tag name)
// - The user navigates to a permalink (i.e. URL with commit ID) for that very same revision
//
// Without additional steps this would result in a second query to resolve the repository information,
// because we now have a different revspec in the URL (the commit ID instead of the symbolic rev).
// But that request would return exactly the same data as the first request. To avoid this, we cache
// the resolved repository information here and reuse if we make a request for a commit ID that we
// have previously seen in a response.
const resolvedRepoRevision = new Map<string, ResolveRepoRevisionResult>()

async function resolveRepoRevision({
    client,
    repoName,
    revspec = '',
}: {
    client: GraphQLClient
    repoName: string
    revspec?: string
}): Promise<ResolvedRevision> {
    const cacheKey = `${repoName}@${revspec}`

    let data: ResolveRepoRevisionResult | undefined

    if (resolvedRepoRevision.has(cacheKey)) {
        // See if we have resolved this revision with another revspec before. This can happen
        // when the user navigates from a symbolic revspec to its respective permalink (commit ID).
        data = resolvedRepoRevision.get(cacheKey)
    } else {
        // See if we have a cached response for the same revision
        data = client.readQuery(ResolveRepoRevision, { repoName, revision: revspec })?.data

        if (shouldResolveRepositoryInformation(data)) {
            data = (
                await client.query(
                    ResolveRepoRevision,
                    { repoName, revision: revspec },
                    { requestPolicy: 'network-only' }
                )
            ).data
        }
    }

    if (!data?.repositoryRedirect) {
        throw new RepoNotFoundError(repoName)
    }

    if (data.repositoryRedirect.__typename === 'Redirect') {
        throw new RepoSeeOtherError(data.repositoryRedirect.url)
    }
    if (data.repositoryRedirect.mirrorInfo.cloneInProgress) {
        throw new CloneInProgressError(repoName, data.repositoryRedirect.mirrorInfo.cloneProgress || undefined)
    }
    if (!data.repositoryRedirect.mirrorInfo.cloned) {
        throw new CloneInProgressError(repoName, 'queued for cloning')
    }

    // The "revision" we queried for could be a commit or a changelist.
    const commit = data.repositoryRedirect.commit || data.repositoryRedirect.changelist?.commit
    if (!commit) {
        throw new RevisionNotFoundError(revspec)
    }

    const defaultBranch = data.repositoryRedirect.defaultBranch?.abbrevName || 'HEAD'

    /*
     * TODO: What exactly is this check for?
    if (!commit.tree) {
        throw new RevisionNotFoundError(defaultBranch)
    }
    */

    // Cache the resolved repository information
    resolvedRepoRevision.set(`${repoName}@${commit.oid}`, data)

    return {
        repo: data.repositoryRedirect,
        commitID: commit.oid,
        defaultBranch,
    }
}

/**
 * We want to resolve the repository and revision information in two cases:
 * - The data is not available yet
 * - The repository is being cloned or the clone is in progress
 *
 * In all other cases, we can use the cached data. That means if the URL specifies a
 * "symbolic" revspec (e.g. a branch or tag name), we will resolve that revspec to the
 * corresponding commit ID only once.
 * This ensures consistentcy as the user navigates to and away from the repository page.
 */
function shouldResolveRepositoryInformation(data: ResolveRepoRevisionResult | undefined): boolean {
    if (!data) {
        return true
    }
    if (data.repositoryRedirect?.__typename === 'Repository') {
        return data.repositoryRedirect.mirrorInfo.cloneInProgress || !data.repositoryRedirect.mirrorInfo.cloned
    }
    return false
}
