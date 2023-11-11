import { FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter, Keyword, Pattern } from '@sourcegraph/shared/src/search/query/token'

enum ValidationErrorType {
    MULTIPLE_REV = 'multiple_rev',
    INVALID_QUERY = 'invalid_query',
    GENERIC_REGEXP = 'generic_regexp',
    HAS_CONTENT_PREDICATE = 'has_content_predicate',
    HAS_FILE_PREDICATE = 'has_file_predicate',
    OR_OPERATOR = 'or_operator',
    AND_OPERATOR = 'and_operator',
}

interface ValidationError {
    type: ValidationErrorType
    reason: string
}

export function validateQueryForExhaustiveSearch(query: string): ValidationError[] {
    const validationErrors: ValidationError[] = []
    const tokens = scanSearchQuery(query)

    if (tokens.type === 'error') {
        validationErrors.push({
            type: ValidationErrorType.INVALID_QUERY,
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
            validationErrors.push({
                type: ValidationErrorType.MULTIPLE_REV,
                reason: 'Multiple rev operators are not compatible',
            })
        }

        const hasTypeFiltersOtherThanFile: boolean = filters
            .filter(filter => resolveFilter(filter.field.value)?.type === FilterType.type && filter.value)
            .some(filter => filter.value?.value !== 'file')

        if (hasTypeFiltersOtherThanFile) {
            validationErrors.push({
                type: ValidationErrorType.INVALID_QUERY,
                reason: 'only type:file is supported',
            })
        }

        const hasRegexpPattern = filters.some(
            filter =>
                resolveFilter(filter.field.value)?.type === FilterType.patterntype && filter.value?.value === 'regexp'
        )
        const hasGenericRegexpPattern = hasRegexpPattern && patterns.some(pattern => pattern.value === '.*')

        if (hasGenericRegexpPattern) {
            validationErrors.push({
                type: ValidationErrorType.GENERIC_REGEXP,
                reason: 'Generic regexp match .* is not compatible',
            })
        }

        const repoHasContentFilter = filters.some(filter => filter.value?.value.startsWith('has.content('))

        if (repoHasContentFilter) {
            validationErrors.push({
                type: ValidationErrorType.HAS_CONTENT_PREDICATE,
                reason: 'repo.has.content predicate is not compatible',
            })
        }

        const repoHasFileFilter = filters.some(filter => filter.value?.value.startsWith('has.file('))

        if (repoHasFileFilter) {
            validationErrors.push({
                type: ValidationErrorType.HAS_FILE_PREDICATE,
                reason: 'repo.has.file predicate is not compatible',
            })
        }

        const hasOr = keywords.some(filter => filter.kind === 'or')

        if (hasOr) {
            validationErrors.push({
                type: ValidationErrorType.OR_OPERATOR,
                reason: 'OR operator is not compatible',
            })
        }

        const hasAnd = keywords.some(filter => filter.kind === 'and')

        if (hasAnd) {
            validationErrors.push({
                type: ValidationErrorType.AND_OPERATOR,
                reason: 'AND operator is not compatible',
            })
        }
    }

    return validationErrors
}
