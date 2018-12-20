import * as clientType from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { FeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

export type ProvideSearchResultSignature = (query: string) => Observable<clientType.SearchResult[] | null | undefined>
export class SearchResultProviderRegistry extends FeatureProviderRegistry<{}, ProvideSearchResultSignature> {
    public provideSearchResult(query: string): Observable<clientType.SearchResult[] | null | undefined> {
        return provideSearchResult(this.providers, query)
    }
}
export function provideSearchResult(
    providers: Observable<ProvideSearchResultSignature[]>,
    query: string,
    logError = true
): Observable<clientType.SearchResult[] | null | undefined> {
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
