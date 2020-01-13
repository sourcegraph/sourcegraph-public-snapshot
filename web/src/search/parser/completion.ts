import * as Monaco from 'monaco-editor'
import { escapeRegExp } from 'lodash'
import { FILTERS, getFilterDefinition } from './filters'
import { Sequence } from './parser'
import { Omit } from 'utility-types'
import { SearchSuggestion } from '../../../../shared/src/graphql/schema'
import { Observable } from 'rxjs'
import { isDefined } from '../../../../shared/src/util/types'

const FILTER_TYPE_COMPLETIONS: Omit<Monaco.languages.CompletionItem, 'range'>[] = FILTERS.flatMap(
    ({ aliases, description }) =>
        aliases.map(
            (label: string): Omit<Monaco.languages.CompletionItem, 'range'> => ({
                label,
                kind: Monaco.languages.CompletionItemKind.Keyword,
                detail: description,
                insertText: `${label}:`,
                filterText: label,
            })
        )
)

/**
 * Returns the completion items for a search query being typed in the Monaco query input,
 * including both static and dynamically fetched suggestions.
 */
export async function getCompletionItems(
    rawQuery: string,
    { members }: Pick<Sequence, 'members'>,
    { column }: Pick<Monaco.Position, 'column'>,
    context: Monaco.languages.CompletionContext,
    fetchSuggestions: (query: string) => Observable<SearchSuggestion[]>
): Promise<Monaco.languages.CompletionList | null> {
    const tokenAtColumn = members.find(({ range }) => range.start + 2 <= column && range.end + 2 >= column)
    if (!tokenAtColumn || tokenAtColumn.token.type === 'whitespace') {
        return null
    }
    const { token, range } = tokenAtColumn
    if (token.type === 'literal') {
        // Offer autocompletion of filter values
        return {
            suggestions: FILTER_TYPE_COMPLETIONS.filter(({ label }) => label.startsWith(token.value)).map(
                (suggestion): Monaco.languages.CompletionItem => ({
                    ...suggestion,
                    range: {
                        startLineNumber: 1,
                        endLineNumber: 1,
                        startColumn: range.start + 1,
                        endColumn: range.end + 2,
                    },
                })
            ),
        }
    }
    if (token.type === 'filter') {
        const { filterValue } = token
        if (!filterValue) {
            return null
        }
        const completingValue = filterValue.range.start + 2 <= column
        if (!completingValue) {
            return null
        }
        const filterDefinition = getFilterDefinition(token.filterType.token.value)
        if (!filterDefinition) {
            return null
        }
        if (filterDefinition.suggestions) {
            const suggestions = await fetchSuggestions(rawQuery).toPromise()
            return {
                suggestions: suggestions
                    .filter(({ __typename }) => __typename === filterDefinition.suggestions)
                    .map((suggestion): Monaco.languages.CompletionItem | null => {
                        if (suggestion.__typename === 'Repository' || suggestion.__typename === 'File') {
                            return {
                                label: suggestion.name,
                                kind: Monaco.languages.CompletionItemKind.Text,
                                insertText: `^${escapeRegExp(
                                    suggestion.__typename === 'File' ? suggestion.path : suggestion.name
                                )}$ `,
                                filterText: `${token.filterType.token.value}:${suggestion.name}`,
                                detail:
                                    suggestion.__typename === 'File'
                                        ? `${suggestion.path} - ${suggestion.repository.name}`
                                        : '',
                                range: {
                                    startLineNumber: 1,
                                    endLineNumber: 1,
                                    startColumn: filterValue.range.start + 1,
                                    endColumn: filterValue.range.end + 2,
                                },
                            }
                        }
                        return null
                    })
                    .filter(isDefined),
            }
        }
        if (filterDefinition.discreteValues) {
            return {
                suggestions: filterDefinition.discreteValues.map(
                    (label): Monaco.languages.CompletionItem => ({
                        label,
                        kind: Monaco.languages.CompletionItemKind.Text,
                        insertText: `${label} `,
                        filterText: label,
                        range: {
                            startLineNumber: 1,
                            endLineNumber: 1,
                            startColumn: filterValue.range.start + 1,
                            endColumn: filterValue.range.end + 2,
                        },
                    })
                ),
            }
        }
    }
    return null
}
