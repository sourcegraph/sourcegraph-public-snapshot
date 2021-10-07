import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { SearchPatternType, SearchVersion } from '../../graphql-operations'
import { firstMatchStreamingSearch, SearchMatch } from '../stream'

export function fetchStreamSuggestions(query: string): Observable<SearchMatch[]> {
    return firstMatchStreamingSearch({
        query,
        version: SearchVersion.V2,
        patternType: SearchPatternType.literal,
        caseSensitive: false,
        versionContext: undefined,
        trace: undefined,
    }).pipe(map(suggestions => suggestions.results))
}
