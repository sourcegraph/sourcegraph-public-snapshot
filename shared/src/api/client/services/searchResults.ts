import * as clientType from '@sourcegraph/extension-api-types'
import { Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { FeatureProviderRegistry } from './registry'

export type ProvideSearchResultSignature = (query: string) => Observable<clientType.SearchResult[] | null | undefined>
export class SearchResultProviderRegistry extends FeatureProviderRegistry<{}, ProvideSearchResultSignature> {
    public provideSearchResult(query: string): Observable<clientType.SearchResult[] | null | undefined> {
        return provideSearchResult(this.providers, query)
    }
}
export function provideSearchResult(
    providers: Observable<ProvideSearchResultSignature[]>,
    query: string
): Observable<clientType.SearchResult[] | null | undefined> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }

            return providers[0](query)
        })
    )
}
