import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import type { PathMatch, RepositoryMatch, SearchMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestionsWithDefaultPatternType } from '@sourcegraph/shared/src/search/suggestions'

export function fetchSuggestions<T extends RepositoryMatch | PathMatch | SymbolMatch, O>(
    query: string,
    patternType: SearchPatternType,
    version: string,
    filterSuggestionFunc: (match: SearchMatch) => match is T,
    mapSuggestionFunc: (match: T) => O
): Observable<O[]> {
    return fetchStreamSuggestionsWithDefaultPatternType(query, patternType, version).pipe(
        map(suggestions => suggestions.filter(filterSuggestionFunc).map(mapSuggestionFunc))
    )
}
