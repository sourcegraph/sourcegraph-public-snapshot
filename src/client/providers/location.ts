import { combineLatest, from, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { Location } from '../../protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

/**
 * Function signature for retrieving related locations given a location (e.g., definition, implementation, and type
 * definition).
 */
export type ProvideTextDocumentLocationSignature<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> = (params: P) => Observable<L | L[] | null>

/** Provides location results from all extensions for definition, implementation, and type definition requests. */
export class TextDocumentLocationProviderRegistry<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> extends FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature<P, L>> {
    public getLocation(params: P): Observable<L | L[] | null> {
        return getLocation<P, L>(this.providers, params)
    }
}

/**
 * Returns an observable that emits all providers' location results whenever any of the last-emitted set of
 * providers emits hovers.
 *
 * Most callers should use the TextDocumentLocationProviderRegistry class, which uses the registered providers.
 */
export function getLocation<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
>(providers: Observable<ProvideTextDocumentLocationSignature<P, L>[]>, params: P): Observable<L | L[] | null> {
    return getLocations(providers, params).pipe(
        map(results => {
            if (results !== null && results.length === 1) {
                return results[0]
            }
            return results
        })
    )
}

/**
 * Like getLocation, except the returned observable never emits singular values, always either an array or null.
 */
export function getLocations<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
>(providers: Observable<ProvideTextDocumentLocationSignature<P, L>[]>, params: P): Observable<L[] | null> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }
            return combineLatest(providers.map(provider => from(provider(params))))
        }),
        map(flattenAndCompact)
    )
}

/**
 * Provides reference results from all extensions.
 *
 * Reference results are always an array or null, unlike results from other location providers (e.g., from
 * textDocument/definition), which can be a single item, an array, or null.
 */
export class TextDocumentReferencesProviderRegistry extends TextDocumentLocationProviderRegistry<ReferenceParams> {
    /** Gets reference locations from all extensions. */
    public getLocation(params: ReferenceParams): Observable<Location[] | null> {
        // References are always an array (unlike other locations, which can be returned as L | L[] |
        // null).
        return getLocations(this.providers, params)
    }
}
