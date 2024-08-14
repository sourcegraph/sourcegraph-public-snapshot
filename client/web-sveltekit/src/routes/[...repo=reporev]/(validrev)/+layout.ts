import { error, redirect } from '@sveltejs/kit'

import type { ResolvedRevision } from '$lib/repo/utils'
import { RevisionNotFoundError, replaceRevisionInURL } from '$lib/shared'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async ({ parent, url }) => {
    // By validating the resolved revision here we can guarantee to
    // subpages that if they load the requested revision exists. This
    // relieves subpages from testing whether the revision is valid.
    const { revision, defaultBranch, resolvedRepository } = await parent()

    const commit = resolvedRepository.commit || resolvedRepository.changelist?.commit

    if (!commit) {
        error(404, new RevisionNotFoundError(revision))
    }

    const isPerforceDepot = !!resolvedRepository.commit?.perforceChangelist

    if (isPerforceDepot && !revision.includes('changelist')) {
        const redirectURL = replaceRevisionInURL(
            url.toString(),
            'changelist/' + resolvedRepository.commit?.perforceChangelist?.cid
        )
        redirect(301, redirectURL)
    }

    return {
        resolvedRevision: {
            repo: resolvedRepository,
            commitID: commit.oid,
            defaultBranch,
        } satisfies ResolvedRevision,
    }
}
