import type { ParamMatcher } from '@sveltejs/kit'

// This ensures that we never consider paths containing /-/ and pointing
// to non-existing pages as repo name
export const match = (param => !param.includes('/-/')) satisfies ParamMatcher
