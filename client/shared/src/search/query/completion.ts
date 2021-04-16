import { escapeRegExp, startCase } from 'lodash'
import * as Monaco from 'monaco-editor'
import { Observable } from 'rxjs'
import { first } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { SymbolKind } from '../../graphql-operations'
import { IRepository, IFile, ISymbol, ILanguage, IRepoGroup, ISearchContext } from '../../graphql/schema'
import { isDefined } from '../../util/types'
import { SearchSuggestion } from '../suggestions'

import { toMonacoRange } from './decoratedToken'
import { FilterType, isNegatableFilter, resolveFilter, FILTERS, escapeSpaces } from './filters'
import { Filter, Token } from './token'

export const repositoryCompletionItemKind = Monaco.languages.CompletionItemKind.Color
const filterCompletionItemKind = Monaco.languages.CompletionItemKind.Customcolor

type PartialCompletionItem = Omit<Monaco.languages.CompletionItem, 'range'>

/**
 * COMPLETION_ITEM_SELECTED is a custom Monaco command that we fire after the user selects an autocomplete suggestion.
 * This allows us to be notified and run custom code when a user selects a suggestion.
 */
export const COMPLETION_ITEM_SELECTED: Monaco.languages.Command = {
    id: 'completionItemSelected',
    title: 'completion item selected',
}

const FILTER_TYPE_COMPLETIONS: Omit<Monaco.languages.CompletionItem, 'range'>[] = Object.keys(FILTERS)
    .flatMap(label => {
        const filterType = label as FilterType
        const completionItem: Omit<Monaco.languages.CompletionItem, 'range' | 'detail'> = {
            label,
            kind: filterCompletionItemKind,
            insertText: `${label}:`,
            filterText: label,
        }
        if (isNegatableFilter(filterType)) {
            return [
                {
                    ...completionItem,
                    detail: FILTERS[filterType].description(false),
                },
                {
                    ...completionItem,
                    label: `-${label}`,
                    insertText: `-${label}:`,
                    filterText: `-${label}`,
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
    // Set a sortText so that filter type suggestions
    // are shown before dynamic suggestions.
    .map((completionItem, index) => ({
        ...completionItem,
        sortText: `0${index}`,
    }))

const repositoryToCompletion = (
    { name }: IRepository,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem => {
    let insertText = options.globbing ? name : `^${escapeRegExp(name)}$`
    insertText = escapeSpaces(insertText)
    insertText = (options.isFilterValue ? insertText : `${FilterType.repo}:${insertText}`) + ' '
    return {
        label: name,
        kind: repositoryCompletionItemKind,
        insertText,
        filterText: name,
        detail: options.isFilterValue ? undefined : 'Repository',
    }
}

const fileToCompletion = (
    { name, path, repository, isDirectory }: IFile,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem => {
    let insertText = options.globbing ? path : `^${escapeRegExp(path)}$`
    insertText = escapeSpaces(insertText)
    insertText = (options.isFilterValue ? insertText : `${FilterType.file}:${insertText}`) + ' '
    return {
        label: name,
        kind: isDirectory ? Monaco.languages.CompletionItemKind.Folder : Monaco.languages.CompletionItemKind.File,
        insertText,
        filterText: name,
        detail: `${path} - ${repository.name}`,
    }
}

/**
 * Maps Sourcegraph SymbolKinds to Monaco CompletionItemKinds.
 */
const symbolKindToCompletionItemKind: Record<SymbolKind, Monaco.languages.CompletionItemKind> = {
    UNKNOWN: Monaco.languages.CompletionItemKind.Value,
    FILE: Monaco.languages.CompletionItemKind.File,
    MODULE: Monaco.languages.CompletionItemKind.Module,
    NAMESPACE: Monaco.languages.CompletionItemKind.Module,
    PACKAGE: Monaco.languages.CompletionItemKind.Module,
    CLASS: Monaco.languages.CompletionItemKind.Class,
    METHOD: Monaco.languages.CompletionItemKind.Method,
    PROPERTY: Monaco.languages.CompletionItemKind.Property,
    FIELD: Monaco.languages.CompletionItemKind.Field,
    CONSTRUCTOR: Monaco.languages.CompletionItemKind.Constructor,
    ENUM: Monaco.languages.CompletionItemKind.Enum,
    INTERFACE: Monaco.languages.CompletionItemKind.Interface,
    FUNCTION: Monaco.languages.CompletionItemKind.Function,
    VARIABLE: Monaco.languages.CompletionItemKind.Variable,
    CONSTANT: Monaco.languages.CompletionItemKind.Constant,
    STRING: Monaco.languages.CompletionItemKind.Value,
    NUMBER: Monaco.languages.CompletionItemKind.Value,
    BOOLEAN: Monaco.languages.CompletionItemKind.Value,
    ARRAY: Monaco.languages.CompletionItemKind.Value,
    OBJECT: Monaco.languages.CompletionItemKind.Value,
    KEY: Monaco.languages.CompletionItemKind.Property,
    NULL: Monaco.languages.CompletionItemKind.Value,
    ENUMMEMBER: Monaco.languages.CompletionItemKind.EnumMember,
    STRUCT: Monaco.languages.CompletionItemKind.Struct,
    EVENT: Monaco.languages.CompletionItemKind.Event,
    OPERATOR: Monaco.languages.CompletionItemKind.Operator,
    TYPEPARAMETER: Monaco.languages.CompletionItemKind.TypeParameter,
}

const symbolToCompletion = ({ name, kind, location }: ISymbol): PartialCompletionItem => ({
    label: name,
    kind: symbolKindToCompletionItemKind[kind],
    insertText: name + ' ',
    filterText: name,
    detail: `${startCase(kind.toLowerCase())} - ${location.resource.repository.name}`,
})

const languageToCompletion = ({ name }: ILanguage): PartialCompletionItem | undefined =>
    name
        ? {
              label: name,
              kind: Monaco.languages.CompletionItemKind.TypeParameter,
              insertText: name + ' ',
              filterText: name,
          }
        : undefined

const repoGroupToCompletion = ({ name }: IRepoGroup): PartialCompletionItem => ({
    label: name,
    kind: repositoryCompletionItemKind,
    insertText: name + ' ',
    filterText: name,
})

const searchContextToCompletion = ({ spec, description }: ISearchContext): PartialCompletionItem => ({
    label: spec,
    kind: repositoryCompletionItemKind,
    insertText: spec + ' ',
    filterText: spec,
    detail: description,
})

const suggestionToCompletionItem = (
    suggestion: SearchSuggestion,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem | undefined => {
    switch (suggestion.__typename) {
        case 'File':
            return fileToCompletion(suggestion, options)
        case 'Repository':
            return repositoryToCompletion(suggestion, options)
        case 'Symbol':
            return symbolToCompletion(suggestion)
        case 'Language':
            return languageToCompletion(suggestion)
        case 'RepoGroup':
            return repoGroupToCompletion(suggestion)
        case 'SearchContext':
            return searchContextToCompletion(suggestion)
    }
}

/**
 * An internal Monaco command causing completion providers to be invoked,
 * and the suggestions widget to be shown.
 *
 * Useful to show the suggestions widget right after selecting a filter type
 * completion, to offer filter values completions.
 */
const TRIGGER_SUGGESTIONS: Monaco.languages.Command = {
    id: 'editor.action.triggerSuggest',
    title: 'Trigger suggestions',
}

const completeStart = (): Monaco.languages.CompletionList => ({
    suggestions: FILTER_TYPE_COMPLETIONS.map(
        (suggestion): Monaco.languages.CompletionItem => ({
            ...suggestion,
            range: {
                startLineNumber: 1,
                endLineNumber: 1,
                startColumn: 1,
                endColumn: 1,
            },
            command: TRIGGER_SUGGESTIONS,
        })
    ),
})

async function completeDefault(
    dynamicSuggestions: Observable<SearchSuggestion[]>,
    token: Token,
    globbing: boolean
): Promise<Monaco.languages.CompletionList> {
    // Offer autocompletion of filter values
    const staticSuggestions = FILTER_TYPE_COMPLETIONS.map(
        (suggestion): Monaco.languages.CompletionItem => ({
            ...suggestion,
            range: toMonacoRange(token.range),
            command: TRIGGER_SUGGESTIONS,
        })
    )
    // If the token being typed matches a known filter,
    // only return static filter type suggestions.
    // This avoids blocking on dynamic suggestions to display
    // the suggestions widget.
    if (
        token.type === 'pattern' &&
        staticSuggestions.some(({ label }) => label.startsWith(token.value.toLowerCase()))
    ) {
        return { suggestions: staticSuggestions }
    }

    return {
        suggestions: [
            ...staticSuggestions,
            ...(await dynamicSuggestions.pipe(first()).toPromise())
                .map(suggestion => suggestionToCompletionItem(suggestion, { isFilterValue: false, globbing }))
                .filter(isDefined)
                .map(completionItem => ({
                    ...completionItem,
                    range: toMonacoRange(token.range),
                    // Set a sortText so that dynamic suggestions
                    // are shown after filter type suggestions.
                    sortText: '1',
                    command: COMPLETION_ITEM_SELECTED,
                })),
        ],
    }
}

async function completeFilter(
    dynamicSuggestions: Observable<SearchSuggestion[]>,
    token: Filter,
    column: number,
    globbing: boolean
): Promise<Monaco.languages.CompletionList | null> {
    const defaultRange = {
        startLineNumber: 1,
        endLineNumber: 1,
        startColumn: column,
        endColumn: column,
    }
    const { value } = token
    const completingValue = !value || value.range.start + 1 <= column
    if (!completingValue) {
        return null
    }
    const resolvedFilter = resolveFilter(token.field.value)
    if (!resolvedFilter) {
        return null
    }
    if (resolvedFilter.definition.suggestions) {
        // If the filter definition has an associated suggestion type,
        // use it to filter dynamic suggestions.
        const suggestions = await dynamicSuggestions.pipe(first()).toPromise()
        return {
            suggestions: suggestions
                .filter(({ __typename }) => __typename === resolvedFilter.definition.suggestions)
                .map(suggestion => suggestionToCompletionItem(suggestion, { isFilterValue: true, globbing }))
                .filter(isDefined)
                .map(partialCompletionItem => ({
                    ...partialCompletionItem,
                    // Set the current value as filterText, so that all dynamic suggestions
                    // returned by the server are displayed. otherwise, if the current filter value
                    // is a regex pattern, Monaco's filtering might hide some suggestions.
                    filterText: value?.value,
                    range: value ? toMonacoRange(value.range) : defaultRange,
                    command: COMPLETION_ITEM_SELECTED,
                })),
        }
    }
    if (resolvedFilter.definition.discreteValues) {
        return {
            suggestions: resolvedFilter.definition.discreteValues(token.value).map(
                (label, index): Monaco.languages.CompletionItem => ({
                    label,
                    sortText: index.toString(), // suggestions sort by order in the list, not alphabetically.
                    kind: Monaco.languages.CompletionItemKind.Value,
                    insertText: `${label} `,
                    filterText: label,
                    range: value ? toMonacoRange(value.range) : defaultRange,
                    command: COMPLETION_ITEM_SELECTED,
                })
            ),
        }
    }
    return null
}

/**
 * Returns the completion items for a search query being typed in the Monaco query input,
 * including both static and dynamically fetched suggestions.
 */
export async function getCompletionItems(
    tokens: Token[],
    { column }: Pick<Monaco.Position, 'column'>,
    dynamicSuggestions: Observable<SearchSuggestion[]>,
    globbing: boolean
): Promise<Monaco.languages.CompletionList | null> {
    if (column === 1) {
        // Show all filter suggestions on the first column.
        return completeStart()
    }
    const tokenAtColumn = tokens.find(({ range }) => range.start + 1 <= column && range.end + 1 >= column)
    if (!tokenAtColumn) {
        throw new Error('getCompletionItems: no token at column')
    }
    const token = tokenAtColumn
    // When the token at column is labeled as a pattern or whitespace, and none of filter,
    // operator, nor quoted value, show static filter type suggestions, followed by dynamic suggestions.
    if (token.type === 'pattern' || token.type === 'whitespace') {
        return completeDefault(dynamicSuggestions, token, globbing)
    }
    if (token.type === 'filter') {
        return completeFilter(dynamicSuggestions, token, column, globbing)
    }
    return null
}
