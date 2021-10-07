import { Remote } from 'comlink'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { FlatExtensionHostAPI } from '../../api/contract'
import { SearchPatternType, SearchVersion } from '../../graphql-operations'
import { firstMatchStreamingSearch, SearchMatch } from '../stream'

export function fetchStreamSuggestions(
    query: string,
    extensionHostAPI: Promise<Remote<FlatExtensionHostAPI>>
): Observable<SearchMatch[]> {
    return firstMatchStreamingSearch({
        query,
        version: SearchVersion.V2,
        patternType: SearchPatternType.literal,
        caseSensitive: false,
        versionContext: undefined,
        trace: undefined,
        extensionHostAPI,
    }).pipe(map(suggestions => suggestions.results))
}
