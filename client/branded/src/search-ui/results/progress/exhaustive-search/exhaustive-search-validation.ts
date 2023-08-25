import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Filter, Keyword, Pattern } from '@sourcegraph/shared/src/search/query/token'

interface ValidationError {
    reason: string
}

export function validateQueryForExhaustiveSearch(query: string): ValidationError[] {
    const validationErrors: ValidationError[] = []
    const tokens = scanSearchQuery(query)

    if (tokens.type === 'error') {
        validationErrors.push({
            reason: `The current query is invalid, problem is at ${tokens.at} column, probably you mean ${tokens.expected}`,
        })
    }

    if (tokens.type === 'success') {
        const filters = tokens.term.filter(token => token.type === 'filter') as Filter[]
        const keywords = tokens.term.filter(token => token.type === 'keyword') as Keyword[]
        const patterns = tokens.term.filter(token => token.type === 'pattern') as Pattern[]

        const hasMultipleRevFilters =
            filters.filter(filter => resolveFilter(filter.field.value)?.type === FilterType.rev && filter.value)
                .length > 1

        if (hasMultipleRevFilters) {
            validationErrors.push({ reason: 'Multiple rev operators are not compatible' })
        }

        const hasRegexpPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype && filter.value?.value === 'regexp'
        )
        const hasGenericRegexpPattern = hasRegexpPattern && patterns.some(pattern => pattern.value === '.*')

        if (hasGenericRegexpPattern) {
            validationErrors.push({ reason: 'Generic regexp match .* is not compatible' })
        }

        const repoHasContentFilter = filters.some(filter => filter.value?.value.startsWith('has.content('))

        if (repoHasContentFilter) {
            validationErrors.push({
                reason: 'repo.has.content predicate is not compatible',
            })
        }

        const repoHasFileFilter = filters.some(filter => filter.value?.value.startsWith('has.file('))

        if (repoHasFileFilter) {
            validationErrors.push({
                reason: 'repo.has.file predicate is not compatible',
            })
        }

        const hasOr = keywords.some(filter => filter.kind === 'or')

        if (hasOr) {
            validationErrors.push({
                reason: 'Or operator is not compatible for exhaustive search',
            })
        }
    }

    return validationErrors
}
