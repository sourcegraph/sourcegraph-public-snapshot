import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { decorate, type Decoration, toDecoration } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'

export function decorateQuery(query: string, searchPatternType?: SearchPatternType): Decoration[] | null {
    const tokens = searchPatternType ? scanSearchQuery(query, false, searchPatternType) : scanSearchQuery(query)
    return tokens.type === 'success'
        ? tokens.term.flatMap(token => decorate(token).map(token => toDecoration(query, token)))
        : null
}
