import { error } from '@sveltejs/kit'

import { isErrorLike } from '$lib/common'

import type { LayoutLoad } from './$types'

export const load: LayoutLoad = async ({ parent }) => {
    // By validating the resolved revision here we can guarantee to
    // subpages that if they load the requested revision exists. This
    // relieves subpages from testing whether the revision is valid.
    const { resolvedRevisionOrError } = await parent()

    if (isErrorLike(resolvedRevisionOrError)) {
        throw error(404, resolvedRevisionOrError)
    }

    return {
        resolvedRevision: resolvedRevisionOrError,
    }
}
