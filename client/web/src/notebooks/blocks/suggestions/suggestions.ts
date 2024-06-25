import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import type { PathMatch, RepositoryMatch, SearchMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestionsVersion } from '@sourcegraph/shared/src/search/suggestions'

export function fetchSuggestions<T extends RepositoryMatch | PathMatch | SymbolMatch, O>(
    query: string,
    version: string,
    filterSuggestionFunc: (match: SearchMatch) => match is T,
    mapSuggestionFunc: (match: T) => O
): Observable<O[]> {
    return fetchStreamSuggestionsVersion(query, version).pipe(
        map(suggestions => suggestions.filter(filterSuggestionFunc).map(mapSuggestionFunc))
    )
}
