import { dirname } from 'path'

import type { Extension } from '@codemirror/state'
import { escapeRegExp } from 'lodash'

import { getRelevantTokens } from '@sourcegraph/shared/src/search/query/analyze'

import { getQueryInformation, type Option, suggestionSources, RenderAs } from '$lib/branded'
import { FilterType, isFilterOfType, type Token } from '$lib/shared'

interface ScopeInformation {
    repoName: string
    revision?: string
    directoryPath?: string
    filePath?: string
    fileLanguage?: string
}

interface ScopeSuggestionsOptions {
    getContextInformation: () => ScopeInformation
}

const EMPTY: any[] = []

function createFilterSuggestion(
    filterType: FilterType,
    filterValue: string,
    description: string,
    tokens: Token[],
    position: number
): Option | null {
    const existingFilter = tokens.find(token => isFilterOfType(token, filterType))
    if (existingFilter && existingFilter.type === 'filter' && existingFilter.value?.value === filterValue) {
        return null
    }

    const label = `${filterType}:${filterValue}`

    return {
        kind: 'context-filter',
        label,
        description,
        action: {
            type: 'completion',
            from: existingFilter ? existingFilter.range.start : position,
            to: existingFilter ? existingFilter.range.end : undefined,
            insertValue: label + ' ',
            name: existingFilter ? 'Replace' : 'Add',
        },
        render: RenderAs.QUERY,
    }
}

/**
 * Returns a suggestion source that provides suggestions relevant to the current repository page.
 * Suggestions include:
 *  - Repository suggestions (e.g., search within the current repository or organization)
 *  - File suggestions (e.g., search within the current file or directory)
 *  - Language suggestions (e.g., search within other files of the same language)
 *
 *  @param options Options to provide context information.
 *  @returns A CodeMirror extension that provides context suggestions.
 */
export function createScopeSuggestions(options: ScopeSuggestionsOptions): Extension {
    const sources = [
        // Creates repo suggestions including prefixes.
        // Example: github.com/sourcegraph/sourcegraph
        //       -> repo:^github\.com/sourcegraph
        //        repo:^github\.com/sourcegraph/sourcegraph$
        // Example: sourcegraph/sourcegraph
        //       -> repo:^sourcegraph
        //          repo:^sourcegraph/sourcegraph$
        (tokens: Token[], position: number, { repoName, revision }: ScopeInformation): Option[] => {
            const options: Option[] = []

            {
                const group = dirname(repoName)
                if (group !== '.') {
                    const option = createFilterSuggestion(
                        FilterType.repo,
                        `^${escapeRegExp(group)}`,
                        'Search within organization/group',
                        tokens,
                        position
                    )
                    if (option) {
                        options.push(option)
                    }
                }
            }

            {
                const option = createFilterSuggestion(
                    FilterType.repo,
                    `^${escapeRegExp(repoName)}$${revision ? `@${revision}` : ''}`,
                    'Search in current repository',
                    tokens,
                    position
                )
                if (option) {
                    options.push(option)
                }
            }

            return options
        },
        // Creates directory and file suggestions, which include the file itself and the directory.
        (tokens: Token[], position: number, { filePath, directoryPath }: ScopeInformation) => {
            const options: Option[] = []

            if (directoryPath && directoryPath !== '.') {
                const option = createFilterSuggestion(
                    FilterType.file,
                    `^${escapeRegExp(directoryPath)}`,
                    'Search in current directory',
                    tokens,
                    position
                )
                if (option) {
                    options.push(option)
                }
            }

            if (filePath) {
                const option = createFilterSuggestion(
                    FilterType.file,
                    `^${escapeRegExp(filePath)}$`,
                    'Search in current file',
                    tokens,
                    position
                )
                if (option) {
                    options.push(option)
                }
            }

            return options
        },
        (tokens: Token[], position: number, { fileLanguage }: ScopeInformation) => {
            if (!fileLanguage) {
                return EMPTY
            }

            const option = createFilterSuggestion(
                FilterType.lang,
                `${fileLanguage}`,
                `Search in other ${fileLanguage} files`,
                tokens,
                position
            )

            return option ? [option] : EMPTY
        },
    ]

    return suggestionSources.of({
        query(state, position) {
            const { parsedQuery, token } = getQueryInformation(state, position)
            // Only show suggestions after whitespace or at the beginning of the query.
            if (token && token.type !== 'whitespace') {
                return null
            }
            const relevantTokens = parsedQuery
                ? getRelevantTokens(
                      parsedQuery,
                      { start: position, end: position },
                      token => token.type === 'parameter'
                  )
                : []
            const context = options.getContextInformation()
            return {
                result: [
                    {
                        title: 'Refine scope',
                        options: sources.flatMap(source => source(relevantTokens, position, context)),
                    },
                ],
            }
        },
    })
}
