import { redirect, error, type Redirect } from '@sveltejs/kit'

import { asError, loadMarkdownSyntaxHighlighting, type ErrorLike } from '$lib/common'
import { resolveRepoRevision, type ResolvedRevision } from '$lib/repo/api/repo'
import { displayRepoName, isRepoSeeOtherErrorLike, isRevisionNotFoundErrorLike, parseRepoRevision } from '$lib/shared'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async ({ params, url, depends }) => {
    // This allows other places to reload all repo related data by calling
    // invalidate('repo:root')
    depends('repo:root')

    // Repo pages render markdown, so we ensure that syntax highlighting for code blocks
    // inside markdown is loaded.
    loadMarkdownSyntaxHighlighting()

    const { repoName, revision } = parseRepoRevision(params.repo)

    let resolvedRevisionOrError: ResolvedRevision | ErrorLike

    try {
        resolvedRevisionOrError = await resolveRepoRevision({ repoName, revision })
    } catch (repoError: unknown) {
        const redirect = isRepoSeeOtherErrorLike(repoError)

        if (redirect) {
            throw redirectToExternalHost(redirect, url)
        }

        // TODO: use differen error codes for different types of errors
        // Let revision errors be handled by the nested layout so that we can
        // still render the main repo navigation and header
        if (!isRevisionNotFoundErrorLike(repoError)) {
            throw error(400, asError(repoError))
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
