import { basename } from 'path'

import {
    autocompletion,
    startCompletion,
    completionKeymap,
    type CompletionResult,
    type Completion,
    snippet,
    type CompletionSource,
    acceptCompletion,
    selectedCompletion,
    currentCompletions,
    setSelectedCompletion,
} from '@codemirror/autocomplete'
import { type Extension, Prec } from '@codemirror/state'
import { keymap, EditorView } from '@codemirror/view'
import {
    mdiCodeArray,
    mdiCodeBraces,
    mdiCodeNotEqual,
    mdiCodeString,
    mdiCube,
    mdiCubeOutline,
    mdiDrawingBox,
    mdiFileDocument,
    mdiFilterOutline,
    mdiFunction,
    mdiHistory,
    mdiKey,
    mdiLightningBoltCircle,
    mdiLink,
    mdiMatrix,
    mdiNull,
    mdiNumeric,
    mdiPackage,
    mdiPiBox,
    mdiPillar,
    mdiPound,
    mdiShape,
    mdiSitemap,
    mdiSourceBranch,
    mdiTextBox,
    mdiTimetable,
    mdiWeb,
    mdiWrench,
} from '@mdi/js'
import { isEqual, startCase } from 'lodash'
import type { NavigateFunction } from 'react-router-dom'

import { isDefined } from '@sourcegraph/common'
import { SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import {
    createFilterSuggestions,
    PREDICATE_REGEX,
    regexInsertText,
    repositoryInsertText,
} from '@sourcegraph/shared/src/search/query/completion-utils'
import { decorate, type DecoratedToken, toDecoration } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { FILTERS, type FilterType, filterTypeKeys, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { getSuggestionQuery } from '@sourcegraph/shared/src/search/query/providers-utils'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import type { SearchMatch } from '@sourcegraph/shared/src/search/stream'
import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'
import { toPrettyBlobURL } from '@sourcegraph/shared/src/util/url'

import { formatRepositoryStarCount } from '../../util'

import { queryTokens } from './parsedQuery'

type CompletionType = SymbolKind | 'queryfilter' | 'repository' | 'searchhistory'

// See SymbolIcon
const typeIconMap: Record<CompletionType, string> = {
    FILE: mdiFileDocument,
    MODULE: mdiCodeBraces,
    NAMESPACE: mdiWeb,
    PACKAGE: mdiPackage,
    CLASS: mdiSitemap,
    METHOD: mdiCubeOutline,
    PROPERTY: mdiWrench,
    FIELD: mdiTextBox,
    CONSTRUCTOR: mdiCubeOutline,
    ENUM: mdiNumeric,
    INTERFACE: mdiLink,
    FUNCTION: mdiFunction,
    VARIABLE: mdiCube,
    CONSTANT: mdiPiBox,
    STRING: mdiCodeString,
    NUMBER: mdiPound,
    BOOLEAN: mdiMatrix,
    ARRAY: mdiCodeArray,
    OBJECT: mdiDrawingBox,
    KEY: mdiKey,
    NULL: mdiNull,
    ENUMMEMBER: mdiNumeric,
    STRUCT: mdiPillar,
    EVENT: mdiTimetable,
    OPERATOR: mdiCodeNotEqual,
    TYPEPARAMETER: mdiCube,
    UNKNOWN: mdiShape,
    queryfilter: mdiFilterOutline,
    repository: mdiSourceBranch,
    searchhistory: mdiHistory,
}

interface SuggestionContext {
    position: number
    onAbort: (listener: () => void) => void
}

/**
 * A suggestion source is given a completion context, the current tokens in the
 * query and the token at the current cursor position. It returns the
 * corresponding completion results.
 * The return type is generic so that it can be used to create different
 * suggestion structures.
 */
type SuggestionSource<R, C extends SuggestionContext> = (
    context: C,
    tokens: Token[],
    tokenAtPosition?: Token
) => R | null | Promise<R | null>

export type StandardSuggestionSource = SuggestionSource<CompletionResult | null, SuggestionContext>

const theme = EditorView.theme({
    '.completion-type-queryfilter > .cm-completionLabel': {
        fontWeight: 'bold',
    },
    '.cm-tooltip-autocomplete svg': {
        width: '1rem',
        height: '1rem',
        display: 'inline-block',
        boxSizing: 'content-box',
        textAlign: 'center',
        paddingRight: '0.5rem',

        '& path': {
            fillOpacity: 0.6,
        },
    },
    '.completion-type-searchhistory > .cm-completionLabel': {
        display: 'none',
    },
    'li.completion-type-searchhistory': {
        height: 'initial !important',
        minHeight: '1.3rem',
    },
})

/**
 * searchQueryAutocompletion registers extensions for automcompletion, using the
 * provided suggestion sources.
 */
export function searchQueryAutocompletion(sources: StandardSuggestionSource[], navigate?: NavigateFunction): Extension {
    const override: CompletionSource[] = sources.map(source => context => {
        const position = context.pos
        const query = context.state.facet(queryTokens)
        const token = query.tokens.find(token => isTokenInRange(token, position))
        return source(
            { position, onAbort: listener => context.addEventListener('abort', listener) },
            query.tokens,
            token
        )
    })

    // Customizing how completion items are rendered
    const addToOptions: NonNullable<Parameters<typeof autocompletion>[0]>['addToOptions'] = [
        // This renders the completion icon
        {
            render(completion) {
                // eslint-disable-next-line @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access
                if ((completion as any)?.url) {
                    return createSVGIcon(mdiLightningBoltCircle, '')
                }
                const icon = createSVGIcon(
                    completion.type && completion.type in typeIconMap
                        ? typeIconMap[completion.type as CompletionType]
                        : typeIconMap[SymbolKind.UNKNOWN],
                    completion.type && completion.type in typeIconMap ? completion.type : ''
                )
                return icon
            },
            // Per CodeMirror documentation, 20 is the default icon
            // position
            position: 20,
        },
        {
            render: voiceOverComma,
            // after icon
            position: 21,
        },
        {
            render: voiceOverComma,
            // after label
            position: 51,
        },
        {
            render(completion) {
                if (completion.type !== 'searchhistory') {
                    return null
                }
                const tokens = scanSearchQuery(completion.label)
                if (tokens.type !== 'success') {
                    throw new Error('this should not happen')
                }
                const nodes = tokens.term
                    .flatMap(token => decorate(token))
                    .map(token => {
                        const decoration = toDecoration(completion.label, token)
                        const node = document.createElement('span')
                        node.className = decoration.className
                        node.textContent = decoration.value
                        return node
                    })

                const container = document.createElement('div')
                container.style.whiteSpace = 'initial'
                for (const node of nodes) {
                    container.append(node)
                }
                return container
            },
            position: 30,
        },
    ]

    return [
        Prec.highest(
            keymap.of([
                ...completionKeymap.map(keybinding => {
                    const { run } = keybinding
                    if (keybinding.key !== 'Enter' || run === undefined) {
                        return keybinding
                    }
                    // Override `Enter` into `Tab` and automatically
                    // accept the first suggestion without an explicit
                    // `DownArrow` to mirror the behavior of the old
                    // "Tab to complete" behavior.
                    return {
                        ...keybinding,
                        key: 'Tab',
                        run(view: EditorView) {
                            if (selectedCompletion(view.state) === null) {
                                // No completion is selected because we
                                // disable the `selectOnOpen` option.
                                if (currentCompletions(view.state).length > 0) {
                                    view.dispatch({ effects: setSelectedCompletion(0) })
                                    acceptCompletion(view)
                                    return true
                                }
                                return false
                            }
                            return run(view)
                        },
                    }
                }),
                {
                    key: 'Enter',
                    run(view) {
                        const selected = selectedCompletion(view.state)
                        // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
                        const url = (selected as any)?.url
                        if (navigate && typeof url === 'string') {
                            navigate(url)
                            return true
                        }
                        // Otherwise apply the selected completion item
                        const hasUserPressedDownArrow = selectedCompletion(view.state) !== null
                        if (hasUserPressedDownArrow) {
                            return acceptCompletion(view)
                        }
                        return false
                    },
                },
            ])
        ),
        theme,
        EditorView.updateListener.of(update => {
            // If a filter was completed, show the completion list again for
            // filter values.
            if (update.transactions.some(transaction => transaction.isUserEvent('input.complete'))) {
                const query = update.state.facet(queryTokens)
                const token = query.tokens.find(token => isTokenInRange(token, update.state.selection.main.anchor - 1))
                if (token) {
                    startCompletion(update.view)
                }
            }
        }),
        autocompletion({
            // We define our own keymap above
            defaultKeymap: false,
            override,
            optionClass: completionItem => 'completion-type-' + (completionItem.type ?? ''),
            icons: false,
            closeOnBlur: true,
            selectOnOpen: false,
            addToOptions,
        }),
    ]
}

export interface DefaultSuggestionSourcesOptions {
    fetchSuggestions: (query: string, onAbort: (listener: () => void) => void) => Promise<SearchMatch[]>
    isSourcegraphDotCom: boolean
    disableFilterCompletion?: true
    disableSymbolCompletion?: true
    showWhenEmpty?: boolean
}

/**
 * Creates default suggestion sources to complete available filters, dynamic
 * suggestions for the current pattern and static and dynamic suggestions for
 * the current filter value.
 */
export function createDefaultSuggestionSources(
    options: DefaultSuggestionSourcesOptions
): SuggestionSource<CompletionResult | null, SuggestionContext>[] {
    const sources: SuggestionSource<CompletionResult | null, SuggestionContext>[] = []

    if (options.disableFilterCompletion !== true) {
        sources.push(
            // Static suggestions shown if the current position is outside a
            // filter value
            createDefaultSource((context, tokens, token) => {
                if (tokens.length === 0) {
                    return null
                }

                // Default to the current cursor position (e.g. if the token is a
                // whitespace, we want the suggestion to be inserted after it)
                let from = context.position

                if (token?.type === 'pattern') {
                    // If the token is a pattern (e.g. the start of a filter name),
                    // we want the suggestion to complete that name.
                    from = token.range.start
                }

                const suggestions = [...FILTER_SUGGESTIONS, ...FILTER_SHORTHAND_SUGGESTIONS].filter(suggestion =>
                    token?.type === 'pattern'
                        ? suggestion.label.toLowerCase().includes(token.value.toLowerCase())
                        : true
                )
                return {
                    from,
                    options: suggestions,
                    // Use filter:false to ensure that static suggestions stay
                    // at the top after dynamic suggestions load. With
                    // `filter:true`, typing a query like `symbol` makes the
                    // `type:symbol` static suggestion appear briefly at the top
                    // until a dynamic suggestion for a repo with a match
                    // against "symbol" gets loaded and the `type:symbol`
                    // suggestion jumps to the bottom.
                    filter: false,
                }
            }),
            // Show static filter value suggestions
            createFilterSource((_context, _tokens, token, resolvedFilter) => {
                if (!resolvedFilter?.definition.discreteValues) {
                    return null
                }

                const { value } = token
                const insidePredicate = value ? PREDICATE_REGEX.test(value.value) : false
                const hasDynamicSuggestions = resolvedFilter.definition.suggestions

                // Don't show static suggestions if we are inside a predicate or
                // if the filter already has a value _and_ is configured for
                // dynamic suggestions.
                // That's because dynamic suggestions are not filtered (filter: false)
                // which CodeMirror always displays above filtered suggestions.
                if (insidePredicate || (value && hasDynamicSuggestions)) {
                    return null
                }

                return {
                    from: value?.range.start ?? token.range.end,
                    to: value?.range.end,
                    // Filtering is unnecessary when dynamic suggestions are
                    // available because if there is any input that the static
                    // suggestions could be filtered by we disable static
                    // suggestions and only show the dynamic ones anyway.
                    filter: !hasDynamicSuggestions,
                    options: resolvedFilter.definition
                        .discreteValues(value, options.isSourcegraphDotCom)
                        .map(({ label, insertText, asSnippet }, index) => {
                            const apply = (insertText || label) + ' '
                            return {
                                label,
                                // See issue https://github.com/sourcegraph/sourcegraph/issues/38254
                                // Per CodeMirror's documentation (https://codemirror.net/docs/ref/#autocomplete.snippet)
                                // "The user can move between fields with Tab and Shift-Tab as long as the fields are
                                // active. Moving to the last field or moving the cursor out of the current field
                                // deactivates the fields."
                                // This means we need to append a field at the end so that pressing Tab when at the last
                                // field will move the cursor after the filter value and not move focus outside the input
                                apply: asSnippet ? snippet(apply + '${}') : apply,
                                // Setting boost this way has the effect of
                                // displaying matching suggestions in the same
                                // order as they have been defined in code.
                                boost: index * -1,
                            }
                        }),
                    // validFor should not be set when the filter has dynamic
                    // suggestions, otherwise static suggestions won't be
                    // removed from the list (because we also disable
                    // filtering above)
                    validFor: hasDynamicSuggestions ? undefined : /^[().:a-z]+$/i,
                }
            }),

            // Show dynamic filter value suggestions
            createFilterSource(async (context, tokens, token, resolvedFilter) => {
                // On Sourcegraph.com, prompt only static suggestions (above) if there is no value to use for generating dynamic suggestions yet.
                if (
                    options.isSourcegraphDotCom &&
                    (!token.value || (token.value.type === 'literal' && token.value.value === ''))
                ) {
                    return null
                }

                if (!resolvedFilter?.definition.suggestions) {
                    return null
                }

                const results: SearchMatch[] = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, resolvedFilter.definition.suggestions),
                    context.onAbort
                )

                if (results.length === 0) {
                    return null
                }

                const filteredResults = results
                    .filter(match => match.type === resolvedFilter.definition.suggestions)
                    .flatMap(match =>
                        completionFromSearchMatch(match, options, token, tokens, { filterValue: token.value?.value })
                    )
                    .filter(isDefined)

                const insidePredicate = token.value ? PREDICATE_REGEX.test(token.value.value) : false

                return {
                    from: token.value?.range.start ?? token.range.end,
                    to: token.value?.range.end,
                    filter: false,
                    options: filteredResults,
                    getMatch: insidePredicate ? undefined : createMatchFunction(token),
                }
            })
        )
    }

    if (options.disableSymbolCompletion !== true) {
        sources.push(
            // Show symbol suggestions outside of filters
            createDefaultSource(async (context, tokens, token) => {
                if (!token || token.type !== 'pattern') {
                    return null
                }

                const results: SearchMatch[] = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, suggestionTypeFromTokens(tokens)),
                    context.onAbort
                )
                if (results.length === 0) {
                    return null
                }

                return {
                    from: token.range.start,
                    to: token.range.end,
                    filter: false,
                    options: results
                        .flatMap(match =>
                            completionFromSearchMatch(match, options, token, tokens, { isDefaultSource: true })
                        )
                        .filter(isDefined),
                }
            })
        )
    }

    return sources
}

// Returns what kind of type to query for based on existing tokens in the query
export function suggestionTypeFromTokens(tokens: Token[]): SearchMatch['type'] {
    let isWithinRepo = false
    let isWithinFile = false
    for (const token of tokens) {
        if (token.type !== 'filter') {
            continue
        }
        switch (token.field.value) {
            case 'type': {
                switch (token.value?.value) {
                    case 'symbol': {
                        return 'symbol'
                    }
                    case 'path':
                    case 'file': {
                        return 'path'
                    }
                    case 'repo': {
                        return 'repo'
                    }
                    case 'diff':
                    case 'commit': {
                        return 'commit'
                    }
                }
                break
            }
            case 'repo':
            case 'r': {
                isWithinRepo = true
                break
            }
            case 'path':
            case 'file':
            case 'f': {
                isWithinFile = true
                break
            }
        }
    }
    // We don't suggest paths because it's easier to get completions for files
    // with the `file:QUERY` filter compared to `type:symbol QUERY`.
    if (isWithinRepo || isWithinFile) {
        return 'symbol'
    }
    return 'repo'
}

type CompletionWithURL = Completion & { url?: string }

function completionFromSearchMatch(
    match: SearchMatch,
    options: DefaultSuggestionSourcesOptions,
    activeToken: Token,
    tokens: Token[],
    params?: {
        filterValue?: string
        isDefaultSource?: boolean
    }
): CompletionWithURL[] {
    const hasNonActivePatternTokens =
        tokens.find(token => token.type === 'pattern' && !isEqual(token.range, activeToken.range)) !== undefined
    switch (match.type) {
        case 'path': {
            return [
                {
                    label: match.path,
                    type: SymbolKind.FILE,
                    url: hasNonActivePatternTokens
                        ? undefined
                        : toPrettyBlobURL({
                              filePath: match.path,
                              revision: match.commit,
                              repoName: match.repository,
                          }),
                    apply: (params?.isDefaultSource ? 'file:' : '') + regexInsertText(match.path) + ' ',
                    info: match.repository,
                },
            ]
        }
        case 'repo': {
            return [
                {
                    label: match.repository,
                    type: 'repository',
                    url: hasNonActivePatternTokens ? undefined : `/${match.repository}`,
                    detail: formatRepositoryStars(match.repoStars),
                    apply: (params?.isDefaultSource ? 'repo:' : '') + repositoryInsertText(match) + ' ',
                },
            ]
        }
        case 'symbol': {
            return match.symbols.map(symbol => ({
                label: (params?.isDefaultSource ? `${symbol.kind.toLowerCase()} ` : '') + symbol.name,
                type: symbol.kind,
                url: hasNonActivePatternTokens ? undefined : symbol.url,
                apply: symbol.name + ' ',
                detail: params?.isDefaultSource
                    ? basename(match.path)
                    : `${startCase(symbol.kind.toLowerCase())} | ${basename(match.path)}`,
                info: match.repository,
            }))
        }
        default: {
            return []
        }
    }
}

function formatRepositoryStars(stars?: number): string {
    const count = formatRepositoryStarCount(stars)
    if (!count) {
        return ''
    }
    return count + ' stars'
}

/**
 * Creates a suggestion source that triggers on no token or pattern or whitespace
 * tokens.
 */
function createDefaultSource<R, C extends SuggestionContext>(source: SuggestionSource<R, C>): SuggestionSource<R, C> {
    return (context, tokens, token) => {
        if (token && token.type !== 'pattern' && token.type !== 'whitespace') {
            return null
        }
        return source(context, tokens, token)
    }
}

type FilterSuggestionSource<R, C extends SuggestionContext> = (
    context: C,
    tokens: Token[],
    filter: Filter,
    resolvedFilter: ReturnType<typeof resolveFilter>
) => ReturnType<SuggestionSource<R, C>>

/**
 * Creates a suggestion source that triggers when a filter value is completed.
 */
function createFilterSource<R, C extends SuggestionContext>(
    source: FilterSuggestionSource<R, C>
): SuggestionSource<R, C> {
    return (context, tokens, token) => {
        // Not completing filter value
        if (!token || token.type !== 'filter' || (token.value && token.value.range.start > context.position)) {
            return null
        }

        const resolvedFilter = resolveFilter(token.field.value)
        if (!resolvedFilter) {
            return null
        }

        return source(context, tokens, token, resolvedFilter)
    }
}

const FILTER_SUGGESTIONS: Completion[] = createFilterSuggestions(Object.keys(FILTERS) as FilterType[]).map(
    ({ label, insertText, detail }) => ({
        label,
        type: 'queryfilter',
        apply: insertText,
        detail,
        boost: insertText.startsWith('-') ? 1 : 2, // demote negated filters
    })
)

// Shorthand suggestions for complete filters like `type:symbol`, `lang:haskell`
// or `patterntype:regexp`.  Shorthand suggestions are helpful for users who may
// not be super familiar with our search syntax and may therefore type "symbol"
// without knowing that the filter is "type:symbol". Without shorthand
// suggestions, the query "symbol" would not have any suggestions because none
// of our filters start with "symbol".
export const FILTER_SHORTHAND_SUGGESTIONS: Completion[] = filterTypeKeys.flatMap(filterType => {
    if (filterType === 'repo' || filterType === 'select') {
        // Ignore shorthand suggestions for repo (because it's noisy) and select
        // (because it has similar values as `type:` suggestions and we assume
        // most users intend to use `type:` instead of `select:`). Feel free to
        // revert this condition in the future if you think repo/select should
        // be included for shorthand suggestions.
        return []
    }
    const completions = FILTERS[filterType].discreteValues?.(undefined, false) ?? []
    return completions.map<Completion>(completion => {
        const insertText = `${filterType}:${completion.label} `.toLowerCase()
        return {
            label: insertText,
            type: 'filter-shorthand',
            insertText,
        }
    })
})

/**
 * This helper function creates a function suitable for CodeMirror's 'getMatch'
 * option. This is used to allow CodeMirror to highlight the matching part of
 * the label.
 * See https://codemirror.net/docs/ref/#autocomplete.CompletionResult.getMatch
 */
function createMatchFunction(token: Filter): ((completion: Completion) => number[]) | undefined {
    if (!token.value?.value) {
        return undefined
    }
    try {
        // Creating a regular expression fails if the value contains special
        // regex characters in invalid positions. In that case we don't
        // highlight.
        const pattern = new RegExp(token.value.value, 'ig')
        return completion => Array.from(completion.label.matchAll(pattern), matchToIndexTuple).flat()
    } catch {
        return undefined
    }
}

/**
 * Converts a regular expression match into an (possibly empty) number tuple
 * representing the start index and the end index of the match.
 */
function matchToIndexTuple(match: RegExpMatchArray): number[] {
    return match.index !== undefined ? [match.index, match.index + match[0].length] : []
}

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function isTokenInRange(
    token: { type: DecoratedToken['type']; range: { start: number; end: number } },
    position: number
): boolean {
    return token.range.start <= position && token.range.end + (token.type === 'field' ? 2 : 0) >= position
}

/**
 * Helper function for inserting voice over-only commas to improve the way the
 * content is read.
 */
function voiceOverComma(): HTMLElement {
    const element = document.createElement('span')
    element.className = 'sr-only'
    element.append(document.createTextNode(','))
    return element
}
