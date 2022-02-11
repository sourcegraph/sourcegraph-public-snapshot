import { getSanitizedRepositories } from '../../../creation-ui-kit'

interface SuggestionsSearchTermInput {
    value: string
    caretPosition: number | null
}

interface SuggestionsSearchTermResult {
    repositories: string[]
    value: string | null
    index: number | null
}

/**
 * Returns search suggestion information.
 * Example:
 * ----------------------- ↓ - caret position --
 * 'github.com/org/about, github.com/org/project
 *
 * Returns index: 1, value: github.com/org/project
 */
export function getSuggestionsSearchTerm(props: SuggestionsSearchTermInput): SuggestionsSearchTermResult {
    const { value, caretPosition } = props

    let startPosition = 0
    // Get repositories array
    const repositories = getSanitizedRepositories(value)

    if (caretPosition === null) {
        return {
            repositories,
            value: null,
            index: null,
        }
    }

    for (const [index, repo] of repositories.entries()) {
        if (startPosition <= caretPosition && startPosition + repo.length >= caretPosition) {
            return {
                repositories,
                index,
                value: repo,
            }
        }

        startPosition += repo.length
    }

    return {
        repositories,
        value: repositories[repositories.length - 1],
        index: repositories.length - 1,
    }
}
