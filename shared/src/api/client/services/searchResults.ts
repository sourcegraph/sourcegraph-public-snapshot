import { Observable } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { GenericSearchResult } from '../../protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'
export type ProvideSearchResultsSignature = (query: string) => Observable<GenericSearchResult[] | null>
export class SearchResultProviderRegistry extends FeatureProviderRegistry<{}, ProvideSearchResultsSignature> {
    public provideSearchResults(query: string): Observable<GenericSearchResult[] | null> {
        return provideSearchResults(this.providers, query)
    }
}
export function provideSearchResults(
    providers: Observable<ProvideSearchResultsSignature[]>,
    query: string
): Observable<GenericSearchResult[] | null> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }
            return providers[0](query)
        })
    )
}
