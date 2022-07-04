import { basename } from 'path'

import {
    autocompletion,
    startCompletion,
    completionKeymap,
    CompletionResult,
    Completion,
    snippet,
    CompletionSource,
} from '@codemirror/autocomplete'
import { Extension, Prec } from '@codemirror/state'
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
    mdiKey,
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
import { startCase } from 'lodash'

import { isDefined } from '@sourcegraph/common'
import { SymbolKind } from '@sourcegraph/search'
import {
    createFilterSuggestions,
    PREDICATE_REGEX,
    regexInsertText,
    repositoryInsertText,
} from '@sourcegraph/shared/src/search/query/completion'
import { DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { FILTERS, FilterType, resolveFilter } from '@sourcegraph/shared/src/search/query/filters'
import { getSuggestionQuery } from '@sourcegraph/shared/src/search/query/providers'
import { Filter, Token } from '@sourcegraph/shared/src/search/query/token'
import { SearchMatch } from '@sourcegraph/shared/src/search/stream'

import { parsedQuery } from './parsedQuery'

import styles from '../CodeMirrorQueryInput.module.scss'

type CompletionType = SymbolKind | 'queryfilter' | 'repository'

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
}

function createIcon(pathSpec: string): Node {
    const svgNS = 'http://www.w3.org/2000/svg'
    const svg = document.createElementNS(svgNS, 'svg')
    svg.setAttributeNS(null, 'viewBox', '0 0 24 24')
    svg.setAttribute('aria-hidden', 'true')

    const path = document.createElementNS(svgNS, 'path')
    path.setAttribute('d', pathSpec)

    svg.append(path)
    return svg
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

/**
 * searchQueryAutocompletion registers extensions for automcompletion, using the
 * provided suggestion sources.
 */
export function searchQueryAutocompletion(
    sources: SuggestionSource<CompletionResult | null, SuggestionContext>[]
): Extension {
    const override: CompletionSource[] = sources.map(source => context => {
        const position = context.pos
        const query = context.state.facet(parsedQuery)
        const token = query.tokens.find(token => isTokenInRange(token, position))
        return source(
            { position, onAbort: listener => context.addEventListener('abort', listener) },
            query.tokens,
            token
        )
    })

    return [
        // Uses the default keymapping but changes accepting suggestions from Enter
        // to Tab
        Prec.highest(
            keymap.of(
                completionKeymap.map(keybinding =>
                    keybinding.key === 'Enter' ? { ...keybinding, key: 'Tab' } : keybinding
                )
            )
        ),
        EditorView.theme({
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
            },
            '.cm-tooltip-autocomplete svg path': {
                fillOpacity: 0.6,
            },
        }),
        EditorView.updateListener.of(update => {
            // If a filter was completed, show the completion list again for
            // filter values.
            if (update.transactions.some(transaction => transaction.isUserEvent('input.complete'))) {
                const query = update.state.facet(parsedQuery)
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
            addToOptions: [
                // This renders the completion icon
                {
                    render(completion) {
                        return createIcon(
                            completion.type && completion.type in typeIconMap
                                ? typeIconMap[completion.type as CompletionType]
                                : typeIconMap[SymbolKind.UNKNOWN]
                        )
                    },
                    // Per CodeMirror documentation, 20 is the default icon
                    // position
                    position: 20,
                },
                // This renders the "Tab" indicator after the details text. It's
                // only visible for the currently selected suggestion (handled
                // by CSS).
                {
                    render() {
                        const node = document.createElement('span')
                        node.className = styles.tabStyle
                        node.textContent = 'Tab'
                        return node
                    },
                    position: 200,
                },
            ],
        }),
    ]
}

/**
 * Creates default suggestion sources to complete available filters, dynamic
 * suggestions for the current pattern and static and dynamic suggestions for
 * the current filter value.
 */
export function createDefaultSuggestionSources(options: {
    fetchSuggestions: (query: string, onAbort: (listener: () => void) => void) => Promise<SearchMatch[]>
    isSourcegraphDotCom: boolean
    globbing: boolean
    disableFilterCompletion?: true
    disableSymbolCompletion?: true
}): SuggestionSource<CompletionResult | null, SuggestionContext>[] {
    const sources: SuggestionSource<CompletionResult | null, SuggestionContext>[] = []

    if (options.disableFilterCompletion !== true) {
        sources.push(
            // Static suggestions shown if the the current position is outside a
            // filter value
            createDefaultSource((context, _tokens, token) => {
                // Default to the current cursor position (e.g. if the token is a
                // whitespace, we want the suggestion to be inserted after it)
                let from = context.position

                if (token?.type === 'pattern') {
                    // If the token is a pattern (e.g. the start of a filter name),
                    // we want the suggestion to complete that name.
                    from = token.range.start
                }

                return {
                    from,
                    options: FILTER_SUGGESTIONS,
                }
            }),
            // Show static filter value suggestions
            createFilterSource((_context, _tokens, token, resolvedFilter) => {
                if (!resolvedFilter?.definition.discreteValues) {
                    return null
                }

                const { value } = token
                const insidePredicate = value ? PREDICATE_REGEX.test(value.value) : false

                if (insidePredicate) {
                    return null
                }

                return {
                    from: value?.range.start ?? token.range.end,
                    options: resolvedFilter.definition
                        .discreteValues(value, options.isSourcegraphDotCom)
                        .map(({ label, insertText, asSnippet }) => {
                            const apply = (insertText || label) + ' '
                            return {
                                label,
                                apply: asSnippet ? snippet(apply) : apply,
                            }
                        }),
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

                const results = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, resolvedFilter.definition.suggestions),
                    context.onAbort
                )
                if (results.length === 0) {
                    return null
                }
                const filteredResults = results
                    .filter(match => match.type === resolvedFilter.definition.suggestions)
                    .map(match => {
                        switch (match.type) {
                            case 'path':
                                return {
                                    label: match.path,
                                    type: SymbolKind.FILE,
                                    apply: regexInsertText(match.path, options) + ' ',
                                    info: match.repository,
                                }
                            case 'repo':
                                return {
                                    label: match.repository,
                                    type: 'repository',
                                    apply:
                                        repositoryInsertText(match, { ...options, filterValue: token.value?.value }) +
                                        ' ',
                                }
                        }
                        return null
                    })
                    .filter(isDefined)

                return {
                    from: token.value?.range.start ?? token.range.end,
                    filter: false,
                    options: filteredResults,
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

                const results = await options.fetchSuggestions(
                    getSuggestionQuery(tokens, token, 'symbol'),
                    context.onAbort
                )
                if (results.length === 0) {
                    return null
                }

                return {
                    from: token.range.start,
                    options: results
                        .flatMap(result => {
                            if (result.type === 'symbol') {
                                const path = result.path
                                return result.symbols.map(symbol => ({
                                    label: symbol.name,
                                    type: symbol.kind,
                                    apply: symbol.name + ' ',
                                    detail: `${startCase(symbol.kind.toLowerCase())} | ${basename(path)}`,
                                    info: result.repository,
                                }))
                            }
                            return null
                        })
                        .filter(isDefined),
                }
            })
        )
    }

    return sources
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

// Looks like there might be a bug with how the end range for a field is
// computed? Need to add 1 to make this work properly.
function isTokenInRange(
    token: { type: DecoratedToken['type']; range: { start: number; end: number } },
    position: number
): boolean {
    return token.range.start <= position && token.range.end + (token.type === 'field' ? 2 : 0) >= position
}
