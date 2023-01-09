import { EditorState, Extension, Facet, StateEffect, StateField } from '@codemirror/state'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { decorate, DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { ParseResult, parseSearchQuery, Node } from '@sourcegraph/shared/src/search/query/parser'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Token } from '@sourcegraph/shared/src/search/query/token'

export interface QueryTokens {
    patternType: SearchPatternType
    tokens: Token[]
}

/**
 * Use this effect to update parse options.
 */
export const setQueryParseOptions = StateEffect.define<{
    patternType: SearchPatternType
    interpretComments?: boolean
}>()

/**
 * Facet representing the parsed query. Other extensions can use this to access
 * the parsed query.
 */
export const queryTokens = Facet.define<QueryTokens, QueryTokens>({
    combine(input) {
        // There will always only be one extension which parses this query
        return input[0] ?? { patternType: SearchPatternType.standard, tokens: [] }
    },
})

export const parsedQuery = Facet.define<ParseResult, Node | null>({
    combine(input) {
        return input[0]?.type === 'success' ? input[0].nodes[0] : null
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
            return [
                queryTokens.compute(['doc', parseOptions], state => {
                    const textDocument = state.sliceDoc()
                    const { patternType, interpretComments } = state.field(parseOptions)
                    if (!textDocument) {
                        return { patternType, tokens: [] }
                    }

                    const result = scanSearchQuery(textDocument, interpretComments, patternType)
                    return {
                        patternType,
                        tokens: result.type === 'success' ? result.term : [],
                    }
                }),
                parsedQuery.compute([queryTokens], state =>
                    parseSearchQuery({ type: 'success', term: state.facet(queryTokens).tokens })
                ),
                decoratedTokens.compute([queryTokens], state => state.facet(queryTokens).tokens.flatMap(decorate)),
            ]
        },
    })
}
