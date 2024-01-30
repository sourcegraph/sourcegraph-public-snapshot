import { redirect, error, type Redirect } from '@sveltejs/kit'

import { asError, loadMarkdownSyntaxHighlighting, type ErrorLike } from '$lib/common'
import type { GraphQLClient } from '$lib/graphql'
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
import { ResolveRepoRevison, ResolvedRepository } from './layout.gql'

export interface ResolvedRevision {
    repo: ResolvedRepository
    commitID: string
    defaultBranch: string
}

export const load: LayoutLoad = async ({ parent, params, url, depends }) => {
    const { graphqlClient: client } = await parent()

    // This allows other places to reload all repo related data by calling
    // invalidate('repo:root')
    depends('repo:root')

    // Repo pages render markdown, so we ensure that syntax highlighting for code blocks
    // inside markdown is loaded.
    loadMarkdownSyntaxHighlighting()

    const { repoName, revision } = parseRepoRevision(params.repo)

    let resolvedRevisionOrError: ResolvedRevision | ErrorLike

    try {
        resolvedRevisionOrError = await resolveRepoRevision({ client, repoName, revision })
    } catch (repoError: unknown) {
        const redirect = isRepoSeeOtherErrorLike(repoError)

        if (redirect) {
            throw redirectToExternalHost(redirect, url)
        }

        // TODO: use differenr error codes for different types of errors
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
        revision,
        resolvedRevisionOrError,
    }
}

function redirectToExternalHost(externalRedirectURL: string, currentURL: URL): Redirect {
    const externalHostURL = new URL(externalRedirectURL)
    const redirectURL = new URL(currentURL)
    // Preserve the path of the current URL and redirect to the repo on the external host.
    redirectURL.host = externalHostURL.host
    redirectURL.protocol = externalHostURL.protocol
    return redirect(303, redirectURL.toString())
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
    let data = client.readQuery({
        query: ResolveRepoRevison,
        variables: {
            repoName,
            revision,
        },
    })
    if (
        !data ||
        (data.repositoryRedirect?.__typename === 'Repository' && data.repositoryRedirect.commit?.oid !== revision)
    ) {
        // We always refetch data when 'revision' is a "symbolic" revision (e.g. a tag or branch name)
        data = await client
            .query({
                query: ResolveRepoRevison,
                variables: {
                    repoName,
                    revision,
                },
                fetchPolicy: 'network-only',
            })
            .then(result => result.data)
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
