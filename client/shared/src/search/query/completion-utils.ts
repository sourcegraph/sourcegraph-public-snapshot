// IMPORTANT: This module contains code used by the CodeMirror query input
// implementation and therefore shouldn't have any runtime dependencies on
// Monaco

import { escapeRegExp } from 'lodash'

import { escapeSpaces, FILTERS, type FilterType, isNegatableFilter } from './filters'

export const PREDICATE_REGEX = /^([.A-Za-z]+)\((.*?)\)?$/

/**
 * regexInsertText escapes the provided value so that it can be used as value
 * for a filter which expects a regular expression.
 */
export const regexInsertText = (value: string): string => {
    const insertText = `^${escapeRegExp(value)}$`
    return escapeSpaces(insertText)
}

/**
 * repositoryInsertText escapes the provides value so that it can be used as a
 * value for the `repo:` filter.
 */
export const repositoryInsertText = ({ repository }: { repository: string }): string => regexInsertText(repository)

/**
 * Given a list of filter types, this function returns a list of objects which
 * can be used for creating completion items. The result also includes negated
 * entries for negateable filters.
 */
export const createFilterSuggestions = (
    filter: FilterType[]
): { label: string; insertText: string; filterText: string; detail: string }[] =>
    filter.flatMap(filterType => {
        const completionItem = {
            label: filterType,
            insertText: `${filterType}:`,
            filterText: filterType,
            detail: '',
        }
        if (isNegatableFilter(filterType)) {
            return [
                {
                    ...completionItem,
                    detail: FILTERS[filterType].description(false),
                },
                {
                    ...completionItem,
                    label: `-${filterType}`,
                    insertText: `-${filterType}:`,
                    filterText: `-${filterType}`,
                    detail: FILTERS[filterType].description(true),
                },
            ]
        }
        return [
            {
                ...completionItem,
                detail: FILTERS[filterType].description,
            },
        ]
    })
