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
        resolvedRevisionOrError = await resolveRepoRevision({ client, repoName, revision })
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

async function resolveRepoRevision({
    client,
    repoName,
    revision = '',
}: {
    client: GraphQLClient
    repoName: string
    revision?: string
}): Promise<ResolvedRevision> {
    // See if we have a cached response
    let data = client.readQuery(ResolveRepoRevision, { repoName, revision })?.data

    if (shouldResolveRepositoryInformation(data)) {
        data = (await client.query(ResolveRepoRevision, { repoName, revision }, { requestPolicy: 'network-only' })).data
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
        throw new RevisionNotFoundError(revision)
    }

    const defaultBranch = data.repositoryRedirect.defaultBranch?.abbrevName || 'HEAD'

    /*
     * TODO: What exactly is this check for?
    if (!commit.tree) {
        throw new RevisionNotFoundError(defaultBranch)
    }
    */

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
