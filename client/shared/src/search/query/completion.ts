import { escapeRegExp, startCase } from 'lodash'
import * as Monaco from 'monaco-editor'
import { Observable } from 'rxjs'
import { first } from 'rxjs/operators'
import { Omit } from 'utility-types'

import { isDefined } from '@sourcegraph/common'

import { SymbolKind } from '../../graphql-operations'
import { MatchedSymbol, PathMatch, RepositoryMatch, SearchMatch } from '../stream'

import { FilterType, isNegatableFilter, resolveFilter, FILTERS, escapeSpaces } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import { Filter, Token } from './token'

export const repositoryCompletionItemKind = Monaco.languages.CompletionItemKind.Color
const filterCompletionItemKind = Monaco.languages.CompletionItemKind.Issue

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
    { repository }: RepositoryMatch,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem => {
    let insertText = options.globbing ? repository : `^${escapeRegExp(repository)}$`
    insertText = escapeSpaces(insertText)
    insertText = (options.isFilterValue ? insertText : `${FilterType.repo}:${insertText}`) + ' '
    return {
        label: repository,
        kind: repositoryCompletionItemKind,
        insertText,
        filterText: repository,
        detail: options.isFilterValue ? undefined : 'Repository',
    }
}

const fileToCompletion = (
    { path, repository }: PathMatch,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem => {
    let insertText = options.globbing ? path : `^${escapeRegExp(path)}$`
    insertText = escapeSpaces(insertText)
    insertText = (options.isFilterValue ? insertText : `${FilterType.file}:${insertText}`) + ' '
    return {
        label: path,
        kind: Monaco.languages.CompletionItemKind.File,
        insertText,
        filterText: path,
        detail: `${path} - ${repository}`,
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

const symbolToCompletion = (
    { name, kind }: MatchedSymbol,
    path: string,
    repository: string
): PartialCompletionItem => ({
    label: name,
    kind: symbolKindToCompletionItemKind[kind],
    insertText: name + ' ',
    filterText: name,
    detail: `${startCase(kind.toLowerCase())} - ${path} - ${repository}`,
})

const suggestionToCompletionItems = (
    suggestion: SearchMatch,
    options: { isFilterValue: boolean; globbing: boolean }
): PartialCompletionItem[] => {
    switch (suggestion.type) {
        case 'path':
            return [fileToCompletion(suggestion, options)]
        case 'repo':
            return [repositoryToCompletion(suggestion, options)]
        case 'symbol':
            return suggestion.symbols.map(symbol => symbolToCompletion(symbol, suggestion.path, suggestion.repository))
    }
    return []
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
    dynamicSuggestions: Observable<SearchMatch[]>,
    token: Token,
    globbing: boolean
): Promise<Monaco.languages.CompletionList> {
    // Offer autocompletion of filter values
    const staticSuggestions = FILTER_TYPE_COMPLETIONS.map(
        (suggestion): Monaco.languages.CompletionItem => ({
            ...suggestion,
            range: toMonacoSingleLineRange(token.range),
            command: TRIGGER_SUGGESTIONS,
        })
    )
    // If the token being typed matches a known filter,
    // only return static filter type suggestions.
    // This avoids blocking on dynamic suggestions to display
    // the suggestions widget.
    if (
        token.type === 'pattern' &&
        staticSuggestions.some(({ label }) => typeof label === 'string' && label.startsWith(token.value.toLowerCase()))
    ) {
        return { suggestions: staticSuggestions }
    }

    return {
        suggestions: [
            ...staticSuggestions,
            ...(await dynamicSuggestions.pipe(first()).toPromise())
                .flatMap(suggestion => suggestionToCompletionItems(suggestion, { isFilterValue: false, globbing }))
                .filter(isDefined)
                .map(completionItem => ({
                    ...completionItem,
                    range: toMonacoSingleLineRange(token.range),
                    // Set a sortText so that dynamic suggestions
                    // are shown after filter type suggestions.
                    sortText: '1',
                    command: COMPLETION_ITEM_SELECTED,
                })),
        ],
    }
}

async function completeFilter(
    serverSuggestions: Observable<SearchMatch[]>,
    token: Filter,
    column: number,
    globbing: boolean,
    isSourcegraphDotCom?: boolean
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
    let staticSuggestions: Monaco.languages.CompletionItem[] = []
    if (resolvedFilter.definition.discreteValues) {
        staticSuggestions = resolvedFilter.definition.discreteValues(token.value, isSourcegraphDotCom).map(
            ({ label, insertText, asSnippet }, index): Monaco.languages.CompletionItem => ({
                label,
                sortText: index.toString().padStart(2, '1'), // suggestions sort by order in the list, not alphabetically (up to 99 values).
                kind: Monaco.languages.CompletionItemKind.Value,
                insertText: `${insertText || label} `,
                filterText: label,
                insertTextRules: asSnippet ? Monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet : undefined,
                range: value ? toMonacoSingleLineRange(value.range) : defaultRange,
                command: COMPLETION_ITEM_SELECTED,
            })
        )
    }
    if (isSourcegraphDotCom === true && (value === undefined || (value.type === 'literal' && value.value === ''))) {
        // On Sourcegraph.com, prompt only static suggestions if there is no value to use for generating dynamic suggestions yet.
        return { suggestions: staticSuggestions }
    }
    let dynamicSuggestions: Monaco.languages.CompletionItem[] = []
    if (resolvedFilter.definition.suggestions) {
        // If the filter definition has an associated dynamic suggestion type,
        // use it to retrieve dynamic suggestions from the backend.
        const suggestions = await serverSuggestions.toPromise()
        dynamicSuggestions = suggestions
            .filter(({ type }) => type === resolvedFilter.definition.suggestions)
            .flatMap(suggestion => suggestionToCompletionItems(suggestion, { isFilterValue: true, globbing }))
            .filter(isDefined)
            .map((partialCompletionItem, index) => ({
                ...partialCompletionItem,
                // Set the current value as filterText, so that all dynamic suggestions
                // returned by the server are displayed. Otherwise, if the current filter value
                // is a regex pattern like `repo:^` Monaco's filtering will try match `^` against the
                // suggestions, and not display them because they don't match.
                filterText: value?.value,
                sortText: index.toString().padStart(2, '0'), // suggestions sort by order in the list, not alphabetically (up to 99 values).
                range: value ? toMonacoSingleLineRange(value.range) : defaultRange,
                command: COMPLETION_ITEM_SELECTED,
            }))
    }
    return { suggestions: staticSuggestions.concat(dynamicSuggestions) }
}

/**
 * Returns the completion items for a search query being typed in the Monaco query input,
 * including both static and dynamically fetched suggestions.
 */
export async function getCompletionItems(
    token: Token,
    { column }: Pick<Monaco.Position, 'column'>,
    dynamicSuggestions: Observable<SearchMatch[]>,
    globbing: boolean,
    isSourcegraphDotCom?: boolean
): Promise<Monaco.languages.CompletionList | null> {
    if (column === 1) {
        // Show all filter suggestions on the first column.
        return completeStart()
    }
    // When the token at column is labeled as a pattern or whitespace, and none of filter,
    // operator, nor quoted value, show static filter type suggestions, followed by dynamic suggestions.
    if (token.type === 'pattern' || token.type === 'whitespace') {
        return completeDefault(dynamicSuggestions, token, globbing)
    }
    if (token.type === 'filter') {
        return completeFilter(dynamicSuggestions, token, column, globbing, isSourcegraphDotCom)
    }
    return null
}
