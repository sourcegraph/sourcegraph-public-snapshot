import { Extension, Facet, StateEffect, StateField } from '@codemirror/state'

import { SearchPatternType } from '@sourcegraph/search'
import { decorate, DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { Token } from '@sourcegraph/shared/src/search/query/token'

export interface ParsedQuery {
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
export const parsedQuery = Facet.define<ParsedQuery, ParsedQuery>({
    combine(input) {
        // There will always only be one extension which parses this query
        return input[0] ?? { patternType: SearchPatternType.literal, tokens: [] }
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

interface ParseOptions {
    patternType: SearchPatternType
    interpretComments?: boolean
}

/**
 * Creates an extension that parses the input as search query and stores the
 * result in the {@link parsedQuery} facet.
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
                parsedQuery.compute(['doc', parseOptions], state => {
                    const { patternType, interpretComments } = state.field(parseOptions)
                    // Looks like Text overwrites toString somehow
                    // eslint-disable-next-line @typescript-eslint/no-base-to-string
                    const result = scanSearchQuery(state.doc.toString(), interpretComments, patternType)
                    return {
                        patternType,
                        tokens: result.type === 'success' ? result.term : [],
                    }
                }),
                decoratedTokens.compute([parsedQuery], state => state.facet(parsedQuery).tokens.flatMap(decorate)),
            ]
        },
    })
}
