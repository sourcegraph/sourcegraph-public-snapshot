import { FilterContext, Result } from './languagespec'

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
