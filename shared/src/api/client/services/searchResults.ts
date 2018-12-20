import * as clientType from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { FeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

export type ProvideSearchResultsSignature = (query: string) => Observable<clientType.SearchResult[] | null>
export class SearchResultProviderRegistry extends FeatureProviderRegistry<{}, ProvideSearchResultsSignature> {
    public provideSearchResults(query: string): Observable<clientType.SearchResult[] | null> {
        return provideSearchResults(this.providers, query)
    }
}
export function provideSearchResults(
    providers: Observable<ProvideSearchResultsSignature[]>,
    query: string,
    logError = true
): Observable<clientType.SearchResult[] | null> {
    return providers.pipe(
        switchMap(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    from(provider(query)).pipe(
                        catchError(err => {
                            if (logError) {
                                console.error(err)
                            }
                            return [null]
                        })
                    )
                )
            )
        ),
        map(flattenAndCompact)
    )
}
