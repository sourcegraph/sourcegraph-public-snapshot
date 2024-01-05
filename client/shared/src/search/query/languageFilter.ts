import { ALL_LANGUAGES, POPULAR_LANGUAGES } from '@sourcegraph/common'

import type { Literal } from './token'

// Returns a list of popular languages initially and a complete list when the
// user has provided input.
export const languageCompletion = (value: Literal | undefined): string[] =>
    value?.value ? ALL_LANGUAGES : POPULAR_LANGUAGES
