import type { ParamMatcher } from '@sveltejs/kit'

import { isKnownSubPage } from '$lib/navigation'

// This ensures that we never consider paths containing /-/ and pointing
// to non-existing pages as repo name
export const match: ParamMatcher = param => {
    // Note: /-/ is a separator between repo revision and repo sub pages
    // Note 2: param doesn't have a leading slash
    return !param.includes('/-/') && !isKnownSubPage('/' + param)
}
