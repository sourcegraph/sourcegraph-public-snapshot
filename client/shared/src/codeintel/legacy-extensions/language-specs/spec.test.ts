import { FilterContext, Result } from './language-spec'

/**
 * A zero-value filter context used for testing.
 */
export const nilFilterContext: FilterContext = {
    repo: '',
    filePath: '',
    fileContent: '',
}

/**
 * A zero-valued search result used for testing.
 */
export const nilResult: Result = {
    repo: '',
    file: '',
}

test('spec.test.ts', () => {})
