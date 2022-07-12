import { escapeRegExp, startCase } from 'lodash'
import * as Monaco from 'monaco-editor'
import { Omit } from 'utility-types'

import { SymbolKind } from '../../graphql-operations'
import { RepositoryMatch, SearchMatch } from '../stream'

import { FilterType, isNegatableFilter, resolveFilter, FILTERS, escapeSpaces, ResolvedFilter } from './filters'
import { toMonacoSingleLineRange } from './monaco'
import { CharacterRange, Filter, Pattern, Token, Whitespace } from './token'

export const repositoryCompletionItemKind = Monaco.languages.CompletionItemKind.Color
const filterCompletionItemKind = Monaco.languages.CompletionItemKind.Issue

type PartialCompletionItem = Omit<Monaco.languages.CompletionItem, 'range'>

export const REPO_DEPS_PREDICATE_REGEX = /^(deps|dependencies|revdeps|dependents)\((.*?)\)?$/
export const PREDICATE_REGEX = /^([.A-Za-z]+)\((.*?)\)?$/

/**
 * COMPLETION_ITEM_SELECTED is a custom Monaco command that we fire after the user selects an autocomplete suggestion.
 * This allows us to be notified and run custom code when a user selects a suggestion.
 */
export const COMPLETION_ITEM_SELECTED: Monaco.languages.Command = {
    id: 'completionItemSelected',
    title: 'completion item selected',
}

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

/**
 * Default filter completions for all filters.
 */
const FILTER_TYPE_COMPLETIONS: Omit<Monaco.languages.CompletionItem, 'range'>[] = createFilterSuggestions(
    Object.keys(FILTERS) as FilterType[]
)
    // Set a sortText so that filter type suggestions
    // are shown before dynamic suggestions.
    .map((completionItem, index) => ({
        ...completionItem,
        kind: filterCompletionItemKind,
        sortText: `0${index}`,
    }))

/**
 * regexInsertText escapes the provided value so that it can be used as value
 * for a filter which expects a regular expression.
 */
export const regexInsertText = (value: string, options: { globbing: boolean }): string => {
    const insertText = options.globbing ? value : `^${escapeRegExp(value)}$`
    return escapeSpaces(insertText)
}

/**
 * repositoryInsertText escapes the provides value so that it can be used as a
 * value for the `repo:` filter.
 */
export const repositoryInsertText = (
    { repository }: RepositoryMatch,
    options: { globbing: boolean; filterValue?: string }
): string => {
    const insertText = regexInsertText(repository, options)

    const depsPredicateMatches = options.filterValue ? options.filterValue.match(REPO_DEPS_PREDICATE_REGEX) : null
    if (depsPredicateMatches) {
        // depsPredicateMatches[1] contains either `deps`, `dependencies`, `revdeps` or `dependents` based on the matched value.
        return `${depsPredicateMatches[1]}(${insertText})`
    }

    return insertText
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

/**
 * Maps a SearchMatch result from the server to a partial competion item for
 * Monaco.
 */
const suggestionToCompletionItems = (
    suggestion: SearchMatch,
    options: { globbing: boolean; filterValue?: string }
): PartialCompletionItem[] | PartialCompletionItem => {
    switch (suggestion.type) {
        case 'path':
            return {
                label: suggestion.path,
                kind: Monaco.languages.CompletionItemKind.File,
                insertText: regexInsertText(suggestion.path, options) + ' ',
                filterText: suggestion.path,
                detail: `${suggestion.path} - ${suggestion.repository}`,
            }
        case 'repo':
            return {
                label: suggestion.repository,
                kind: repositoryCompletionItemKind,
                insertText: repositoryInsertText(suggestion, options) + ' ',
                filterText: suggestion.repository,
            }
        case 'symbol':
            return suggestion.symbols.map(({ name, kind }) => ({
                label: name,
                kind: symbolKindToCompletionItemKind[kind],
                insertText: name + ' ',
                filterText: name,
                detail: `${startCase(kind.toLowerCase())} - ${suggestion.path} - ${suggestion.repository}`,
            }))
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

/**
 * completeDefault returns completions for filters or symbols.
 * If the query is empty (no token is availabe) or if completion is triggered at
 * a whitespace, it returns a static list of available filters.
 * If completion is triggered at a pattern and the pattern does not match a
 * known filter, it returns symbol suggestions from the server.
 */
async function completeDefault(
    tokenAtPosition: Whitespace | Pattern | null,
    serverSuggestions: (token: Token, type: SearchMatch['type']) => Promise<PartialCompletionItem[]>,
    disablePatternSuggestions = false
): Promise<Monaco.languages.CompletionList> {
    // Default filter suggestions
    let suggestions = FILTER_TYPE_COMPLETIONS.map(
        (suggestion): Monaco.languages.CompletionItem => ({
            ...suggestion,
            range: toMonacoSingleLineRange(tokenAtPosition?.range || startRange),
            command: TRIGGER_SUGGESTIONS,
        })
    )

    if (tokenAtPosition?.type === 'pattern') {
        // If the token being typed matches a known filter, only return static
        // filter type suggestions.
        // This avoids blocking on dynamic suggestions to display the
        // suggestions widget.
        if (
            suggestions.some(
                ({ label }) => typeof label === 'string' && label.startsWith(tokenAtPosition.value.toLowerCase())
            )
        ) {
            return { suggestions }
        }

        if (!disablePatternSuggestions) {
            suggestions = suggestions.concat(
                (await serverSuggestions(tokenAtPosition, 'symbol')).map(completionItem => ({
                    ...completionItem,
                    range: toMonacoSingleLineRange(tokenAtPosition.range),
                    // Set a sortText so that dynamic suggestions
                    // are shown after filter type suggestions.
                    sortText: '1',
                    command: COMPLETION_ITEM_SELECTED,
                }))
            )
        }
    }

    return { suggestions }
}

/**
 * completeFilterValue returns suggestions for filters with static values (e.g.
 * "case") or, if the filter is configured as such, fetches suggestions from the
 * server.
 */
async function completeFilterValue(
    resolvedFilter: NonNullable<ResolvedFilter>,
    token: Filter,
    serverSuggestions: (token: Filter, type: SearchMatch['type']) => Promise<PartialCompletionItem[]>,
    column: number,
    isSourcegraphDotCom?: boolean
): Promise<Monaco.languages.CompletionItem[]> {
    const defaultRange = {
        startLineNumber: 1,
        endLineNumber: 1,
        startColumn: column,
        endColumn: column,
    }
    const { value } = token

    // FIXME(tsenart): We need to refactor completions to work with
    // complex predicates like repo:dependencies()
    // For now we just disable all static suggestions for a predicate's filter
    // if we are inside that predicate.
    const insidePredicate = value ? PREDICATE_REGEX.test(value.value) : false

    let suggestions: Monaco.languages.CompletionItem[] = []
    if (resolvedFilter.definition.discreteValues && !insidePredicate) {
        suggestions = resolvedFilter.definition.discreteValues(token.value, isSourcegraphDotCom).map(
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
        return suggestions
    }
    if (resolvedFilter.definition.suggestions) {
        // If the filter definition has an associated dynamic suggestion type,
        // use it to retrieve dynamic suggestions from the backend.
        suggestions = suggestions.concat(
            (await serverSuggestions(token, resolvedFilter.definition.suggestions)).map(
                (partialCompletionItem, index) => ({
                    ...partialCompletionItem,
                    // Set the current value as filterText, so that all dynamic suggestions
                    // returned by the server are displayed. Otherwise, if the current filter value
                    // is a regex pattern like `repo:^` Monaco's filtering will try match `^` against the
                    // suggestions, and not display them because they don't match.
                    filterText: value?.value,
                    sortText: index.toString().padStart(2, '0'), // suggestions sort by order in the list, not alphabetically (up to 99 values).
                    range: value ? toMonacoSingleLineRange(value.range) : defaultRange,
                    command: COMPLETION_ITEM_SELECTED,
                })
            )
        )
    }
    return suggestions
}

const startRange: CharacterRange = { start: 1, end: 1 }

export type FetchSuggestions = <T extends SearchMatch['type']>(
    token: Token,
    type: T
) => Promise<Extract<SearchMatch, { type: T }>[]>

/**
 * Returns the completion items for a search query being typed in the Monaco query input,
 * including both static and dynamically fetched suggestions.
 */
export async function getCompletionItems(
    tokenAtPosition: Token | null,
    { column }: Pick<Monaco.Position, 'column'>,
    fetchDynamicSuggestions: FetchSuggestions,
    {
        globbing = false,
        isSourcegraphDotCom = false,
        disablePatternSuggestions = false,
    }: { globbing?: boolean; isSourcegraphDotCom?: boolean; disablePatternSuggestions?: boolean }
): Promise<Monaco.languages.CompletionList | null> {
    let suggestions: Monaco.languages.CompletionItem[] = []

    // Show all filter suggestions if the query is empty or when the token at
    // column is labeled as a pattern or whitespace, and none of filter,
    // operator, nor quoted value, show static filter type suggestions, followed
    // by dynamic suggestions.
    if (!tokenAtPosition || tokenAtPosition.type === 'pattern' || tokenAtPosition.type === 'whitespace') {
        return completeDefault(
            tokenAtPosition,
            async (token, type) =>
                (await fetchDynamicSuggestions(token, type)).flatMap(suggestion =>
                    suggestionToCompletionItems(suggestion, { globbing })
                ),
            disablePatternSuggestions
        )
    }

    if (tokenAtPosition?.type === 'filter') {
        const completingValue = !tokenAtPosition.value || tokenAtPosition.value.range.start + 1 <= column
        if (!completingValue) {
            return null
        }

        const resolvedFilter = resolveFilter(tokenAtPosition.field.value)
        if (!resolvedFilter) {
            return null
        }

        suggestions = suggestions.concat(
            await completeFilterValue(
                resolvedFilter,
                tokenAtPosition,
                async (token, type) =>
                    (await fetchDynamicSuggestions(token, type)).flatMap(suggestion =>
                        suggestionToCompletionItems(suggestion, { globbing, filterValue: token.value?.value })
                    ),
                column,
                isSourcegraphDotCom
            )
        )
        return { suggestions }
    }

    return null
}
