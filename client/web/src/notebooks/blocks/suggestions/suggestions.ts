import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { PathMatch, RepositoryMatch, SearchMatch, SymbolMatch } from '@sourcegraph/shared/src/search/stream'
import { fetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'

export function fetchSuggestions<T extends RepositoryMatch | PathMatch | SymbolMatch, O>(
    query: string,
    filterSuggestionFn: (match: SearchMatch) => match is T,
    mapSuggestionFn: (match: T) => O
): Observable<O[]> {
    return fetchStreamSuggestions(query).pipe(
        map(suggestions => suggestions.filter(filterSuggestionFn).map(mapSuggestionFn))
    )
}
