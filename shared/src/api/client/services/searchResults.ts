import { Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { SearchResult } from '../../protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'
export type ProvideSearchResultSignature = (query: string) => Observable<SearchResult[] | null>
export class SearchResultProviderRegistry extends FeatureProviderRegistry<{}, ProvideSearchResultSignature> {
    public provideSearchResults(query: string): Observable<SearchResult[] | null> {
        return provideSearchResults(this.providers, query)
    }
}
export function provideSearchResults(
    providers: Observable<ProvideSearchResultSignature[]>,
    query: string
): Observable<SearchResult[] | null> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }
            return providers[0](query)
        })
    )
}
