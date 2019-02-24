import { Observable, of } from 'rxjs'
import { mergeMap, switchMap } from 'rxjs/operators'
import { FeatureProviderRegistry } from './registry'

export type TransformQuerySignature = (query: string) => Observable<string>

/** Transforms search queries using registered query transformers from extensions. */
export class QueryTransformerRegistry extends FeatureProviderRegistry<{}, TransformQuerySignature> {
    public transformQuery(query: string): Observable<string> {
        return transformQuery(this.providers, query)
    }
}

/**
 * Returns an observable that emits a query transformed by all providers whenever any of the last-emitted set of providers emits
 * a query.
 *
 * Most callers should use QueryTransformerRegistry's transformQuery method, which uses the registered query transformers
 *
 */
export function transformQuery(providers: Observable<TransformQuerySignature[]>, query: string): Observable<string> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [query]
            }
            return providers.reduce<Observable<string>>(
                (currentQuery, transformQuery) => currentQuery.pipe(mergeMap(q => transformQuery(q))),
                of(query)
            )
        })
    )
}
