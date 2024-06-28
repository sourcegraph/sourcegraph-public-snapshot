import { type EditorState, type Extension, Facet, StateEffect, StateField } from '@codemirror/state'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { decorate, type DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { type ParseResult, parseSearchQuery, type Node } from '@sourcegraph/shared/src/search/query/parser'
import { detectPatternType, scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import type { Filter, Token } from '@sourcegraph/shared/src/search/query/token'

export interface QueryTokens {
    patternType: SearchPatternType
    tokens: Token[]
}

export const parsedQuery = Facet.define<ParseResult, Node | null>({
    combine(input) {
        return input[0]?.type === 'success' ? input[0].node : null
    },
})

/**
 * Facet representing decorated tokens (which includes e.g. regular
 * expressions).
 */
export const decoratedTokens = Facet.define<DecoratedToken[], DecoratedToken[]>({
    combine(input) {
        return input[0] ?? []
    },
})

/**
 * Facet representing the parsed query. Other extensions can use this to access
 * the parsed query. Populates the {@link parsedQuery} and {@link decoratedTokens} facets.
 */
export const queryTokens: Facet<QueryTokens, QueryTokens> = Facet.define<QueryTokens, QueryTokens>({
    combine(input) {
        // There will always only be one extension which parses this query
        return input[0] ?? { patternType: SearchPatternType.standard, tokens: [] }
    },
    enables(self) {
        return [
            parsedQuery.compute([self], state => parseSearchQuery({ type: 'success', term: state.facet(self).tokens })),
            decoratedTokens.compute([self], state => state.facet(self).tokens.flatMap(decorate)),
        ]
    },
})

interface ParseOptions {
    patternType: SearchPatternType
    interpretComments?: boolean
}

/**
 * Creates an extension that parses the input as search query and stores the
 * result in the {@link queryTokens} facet.
 */
export function parseInputAsQuery(initialParseOptions: ParseOptions): Extension {
    // Editor state to keep information about how to parse the query. Can be updated
    // with the `setQueryParseOptions` effect.
    return StateField.define<ParseOptions>({
        create() {
            return initialParseOptions
        },
        update(value, transaction) {
            for (const effect of transaction.effects) {
                if (effect.is(setQueryParseOptions)) {
                    return { ...value, ...effect.value }
                }
            }
            return value
        },
        provide(parseOptions) {
            // Parse the query using our existing parser. It depends on the
            // current input (obviously) and the current parse options. It gets
            // recomputed whenever one of those values changes.
            return queryTokens.compute(['doc', parseOptions], state => {
                const textDocument = state.sliceDoc()
                const options = state.field(parseOptions)
                if (!textDocument) {
                    return { patternType: options.patternType, tokens: [] }
                }

                const patternType = detectPatternType(textDocument) || options.patternType
                const result = scanSearchQuery(textDocument, options.interpretComments, patternType)

                return {
                    patternType,
                    tokens: result.type === 'success' ? result.term : [],
                }
            })
        },
    })
}

/**
 * Use this effect to update parse options.
 */
export const setQueryParseOptions = StateEffect.define<{
    patternType: SearchPatternType
    interpretComments?: boolean
}>()

/**
 * Returns the token at the current position (if any)
 */
export function tokenAt(tokens: Token[], position: number): Token | undefined {
    // We do a exclusive end check for whitespace tokens so that the token that
    // possibly follows the whitespace token is picked instead.
    return tokens.find(({ range, type }) =>
        range.start <= position && type === 'whitespace' ? range.end > position : range.end >= position
    )
}

/**
 * Returns the current query tokens
 */
export function tokens(state: EditorState): Token[] {
    return state.facet(queryTokens).tokens
}

export function getParsedQuery(state: EditorState): Node | null {
    return state.facet(parsedQuery)
}

/**
 * Returns the parsed query and the token at the provided position.
 * The list of tokens returned by this function is pre-processed to
 * handle open strings better.
 */
export function getQueryInformation(
    state: EditorState,
    position: number
): { parsedQuery: Node | null; tokens: Token[]; token: Token | undefined; input: string; position: number } {
    const input = state.sliceDoc()
    const queryTokens = collapseOpenFilterValues(tokens(state), input)

    return {
        parsedQuery: getParsedQuery(state),
        tokens: queryTokens,
        token: tokenAt(queryTokens, position),
        input,
        position,
    }
}

/**
 * Helper function to convert filter values that start with a quote but are not
 * closed yet (e.g. author:"firstname lastna|) to a single filter token to
 * prevent irrelevant suggestions.
 */
export function collapseOpenFilterValues(tokens: Token[], input: string): Token[] {
    const result: Token[] = []
    let openFilter: Filter | null = null
    let hold: Token[] = []

    function mergeFilter(filter: Filter, values: Token[]): Filter {
        if (!filter.value?.value) {
            // For simplicity but this should never occure
            return filter
        }
        const end = values.at(-1)?.range.end ?? filter.value.range.end
        return {
            ...filter,
            range: {
                start: filter.range.start,
                end,
            },
            value: {
                ...filter.value,
                range: {
                    start: filter.value.range.start,
                    end,
                },
                value:
                    filter.value.value + values.map(token => input.slice(token.range.start, token.range.end)).join(''),
            },
        }
    }

    for (const token of tokens) {
        switch (token.type) {
            case 'filter': {
                {
                    if (token.value?.value.startsWith('"') && !token.value.quoted) {
                        openFilter = token
                    } else {
                        if (openFilter?.value) {
                            result.push(mergeFilter(openFilter, hold))
                            openFilter = null
                            hold = []
                        }
                        result.push(token)
                    }
                }
                break
            }
            case 'pattern':
            case 'whitespace': {
                if (openFilter) {
                    hold.push(token)
                } else {
                    result.push(token)
                }
                break
            }
            default: {
                if (openFilter?.value) {
                    result.push(mergeFilter(openFilter, hold))
                    openFilter = null
                    hold = []
                }
                result.push(token)
            }
        }
    }

    if (openFilter?.value) {
        result.push(mergeFilter(openFilter, hold))
    }

    return result
}
