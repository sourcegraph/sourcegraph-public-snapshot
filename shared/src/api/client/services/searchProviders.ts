import { EMPTY, from, merge, Observable, ObservableInput, of } from 'rxjs'
import { catchError, defaultIfEmpty, filter, map } from 'rxjs/operators'
import { SearchOptions, SearchQuery, TextSearchResult } from 'sourcegraph'
import { FeatureProviderRegistry } from './registry'

export interface ProvideTextSearchResultsParams {
    query: SearchQuery
    options?: SearchOptions
}

export type ProvideTextSearchResultsSignature = (
    params: ProvideTextSearchResultsParams
) => ObservableInput<TextSearchResult[]>

/**
 * Provides search results from matching registered providers.
 */
export class SearchProviderRegistry extends FeatureProviderRegistry<{}, ProvideTextSearchResultsSignature> {
    /**
     * Gets results from all registered serach providers. Returns an observable that, initially and
     * upon the providers changing, emits an observable of all providers' results.
     *
     * Using a higher-order observable here lets the caller display a loading indicator when the
     * inner observable has not yet completed. The outer observable never completes because
     * providers may be registered and unregistered at any time.
     */
    public getResults(params: ProvideTextSearchResultsParams): Observable<Observable<TextSearchResult[]>> {
        return getResults(this.providers, params)
    }
}

/**
 * Returns the combined results of invoking multiple search providers.
 *
 * @internal Callers should instead use {@link SearchProviderRegistry#getResults}.
 */
export function getResults(
    providers: Observable<ProvideTextSearchResultsSignature[]>,
    params: ProvideTextSearchResultsParams,
    logErrors = true
): Observable<Observable<TextSearchResult[]>> {
    return providers.pipe(
        map(providers =>
            providers.length > 0
                ? merge(
                      ...providers.map(provider =>
                          from(provider(params)).pipe(
                              catchError(err => {
                                  if (logErrors) {
                                      console.error(err)
                                  }
                                  return EMPTY
                              }),
                              filter(results => results.length > 0)
                          )
                      )
                  ).pipe(defaultIfEmpty([] as TextSearchResult[]))
                : of([])
        )
    )
}
