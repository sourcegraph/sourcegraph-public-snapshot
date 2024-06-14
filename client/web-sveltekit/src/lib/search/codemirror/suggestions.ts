import { dirname } from 'path'

import type { Extension } from '@codemirror/state'
import { mdiFileOutline, mdiFilterOutline, mdiFolderOutline, mdiSourceRepository } from '@mdi/js'
import { escapeRegExp } from 'lodash'

import { getQueryInformation, type Option, suggestionSources, RenderAs } from '$lib/branded'
import {
    EMPTY_RELEVANT_TOKEN_RESULT,
    FilterType,
    getRelevantTokens,
    isFilterOfType,
    type RelevantTokenResult,
} from '$lib/shared'

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
    result: RelevantTokenResult,
    position: number,
    icon: string
): Option | null {
    const existingFilter = result.tokens.find(token => isFilterOfType(token, filterType))
    if (existingFilter && existingFilter.type === 'filter' && existingFilter.value?.value === filterValue) {
        return null
    }

    const existingRange = existingFilter ? result.sourceMap.get(existingFilter) : undefined
    const label = `${filterType}:${filterValue}`

    return {
        kind: 'context-filter',
        label,
        description,
        action: {
            type: 'completion',
            from: existingRange ? existingRange.start : position,
            to: existingRange ? existingRange.end : undefined,
            insertValue: label + ' ',
            name: existingRange ? 'Replace' : 'Add',
        },
        icon,
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
        (result: RelevantTokenResult, position: number, { repoName, revision }: ScopeInformation): Option[] => {
            const options: Option[] = []

            {
                let group = dirname(repoName)
                if (group !== '.') {
                    if (!group.endsWith('/')) {
                        group += '/'
                    }
                    const option = createFilterSuggestion(
                        FilterType.repo,
                        `^${escapeRegExp(group)}`,
                        'Search within organization/group',
                        result,
                        position,
                        mdiSourceRepository
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
                    result,
                    position,
                    mdiSourceRepository
                )
                if (option) {
                    options.push(option)
                }
            }

            return options
        },
        // Creates directory and file suggestions, which include the file itself and the directory.
        (result: RelevantTokenResult, position: number, { filePath, directoryPath }: ScopeInformation) => {
            const options: Option[] = []

            if (directoryPath && directoryPath !== '.') {
                if (!directoryPath.endsWith('/')) {
                    directoryPath += '/'
                }
                const option = createFilterSuggestion(
                    FilterType.file,
                    `^${escapeRegExp(directoryPath)}`,
                    'Search in current directory',
                    result,
                    position,
                    mdiFolderOutline
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
                    result,
                    position,
                    mdiFileOutline
                )
                if (option) {
                    options.push(option)
                }
            }

            return options
        },
        (result: RelevantTokenResult, position: number, { fileLanguage }: ScopeInformation) => {
            if (!fileLanguage) {
                return EMPTY
            }

            const option = createFilterSuggestion(
                FilterType.lang,
                `${fileLanguage}`,
                `Search in other ${fileLanguage} files`,
                result,
                position,
                mdiFilterOutline
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
                : EMPTY_RELEVANT_TOKEN_RESULT
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
