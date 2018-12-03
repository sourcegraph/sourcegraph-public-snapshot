import { combineLatest, from, Observable, of } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { ReferenceParams, TextDocumentPositionParams } from '../../protocol'
import { Location } from '../../protocol/plainTypes'
import { Model, modelToTextDocumentPositionParams } from '../model'
import { match } from '../types/textDocument'
import { DocumentFeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

/**
 * Function signature for retrieving related locations given a location (e.g., definition, implementation, and type
 * definition).
 */
export type ProvideTextDocumentLocationSignature<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> = (params: P) => Observable<L | L[] | null>

/**
 * Provides location results from matching registered providers for definition, implementation, and type definition
 * requests.
 */
export class TextDocumentLocationProviderRegistry<
    P extends TextDocumentPositionParams = TextDocumentPositionParams,
    L extends Location = Location
> extends DocumentFeatureProviderRegistry<ProvideTextDocumentLocationSignature<P, L>> {
    public getLocation(params: P): Observable<L | L[] | null> {
        return getLocation<P, L>(this.providersForDocument(params.textDocument), params)
    }

    public getLocationsAndProviders(
        model: Observable<Pick<Model, 'visibleViewComponents'>>,
        extraParams?: Pick<P, Exclude<keyof P, keyof TextDocumentPositionParams>>
    ): Observable<{ locations: Observable<Location[] | null> | null; hasProviders: boolean }> {
        return combineLatest(this.entries, model).pipe(
            map(([entries, { visibleViewComponents }]) => {
                const params = modelToTextDocumentPositionParams({ visibleViewComponents })
                if (!params) {
                    return { locations: null, hasProviders: false }
                }

                const providers = entries
                    .filter(({ registrationOptions }) =>
                        match(registrationOptions.documentSelector, params.textDocument)
                    )
                    .map(({ provider }) => provider)
                return {
                    locations: getLocations<P, L>(of(providers), {
                        ...(params as Pick<P, keyof TextDocumentPositionParams>),
                        ...(extraParams as Pick<P, Exclude<keyof P, 'textDocument' | 'position'>>),
                    } as P),
                    hasProviders: providers.length > 0,
                }
            })
        )
    }
}

/**
 * Returns an observable that emits the providers' location results whenever any of the last-emitted set of
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
 * Provides reference results from all providers.
 *
 * Reference results are always an array or null, unlike results from other location providers (e.g., from
 * textDocument/definition), which can be a single item, an array, or null.
 */
export class TextDocumentReferencesProviderRegistry extends TextDocumentLocationProviderRegistry<ReferenceParams> {
    /** Gets reference locations from all matching providers. */
    public getLocation(params: ReferenceParams): Observable<Location[] | null> {
        // References are always an array (unlike other locations, which can be returned as L | L[] |
        // null).
        return getLocations(this.providersForDocument(params.textDocument), params)
    }
}
