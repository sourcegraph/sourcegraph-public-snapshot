import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import type { PathMatch, RepositoryMatch, SearchMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestionsPatternTypeVersion } from '@sourcegraph/shared/src/search/suggestions'

export function fetchSuggestions<T extends RepositoryMatch | PathMatch | SymbolMatch, O>(
    query: string,
    patternType: SearchPatternType,
    filterSuggestionFunc: (match: SearchMatch) => match is T,
    mapSuggestionFunc: (match: T) => O
): Observable<O[]> {
    return fetchStreamSuggestionsPatternTypeVersion(query, patternType).pipe(
        map(suggestions => suggestions.filter(filterSuggestionFunc).map(mapSuggestionFunc))
    )
}
