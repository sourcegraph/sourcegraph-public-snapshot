import { Observable } from 'rxjs';
import { ReferenceParams, TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol';
import { Location } from '../../protocol/plainTypes';
import { FeatureProviderRegistry } from './registry';
/**
 * Function signature for retrieving related locations given a location (e.g., definition, implementation, and type
 * definition).
 */
export declare type ProvideTextDocumentLocationSignature<P extends TextDocumentPositionParams = TextDocumentPositionParams, L extends Location = Location> = (params: P) => Observable<L | L[] | null>;
/** Provides location results from all extensions for definition, implementation, and type definition requests. */
export declare class TextDocumentLocationProviderRegistry<P extends TextDocumentPositionParams = TextDocumentPositionParams, L extends Location = Location> extends FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentLocationSignature<P, L>> {
    getLocation(params: P): Observable<L | L[] | null>;
}
/**
 * Returns an observable that emits all providers' location results whenever any of the last-emitted set of
 * providers emits hovers.
 *
 * Most callers should use the TextDocumentLocationProviderRegistry class, which uses the registered providers.
 */
export declare function getLocation<P extends TextDocumentPositionParams = TextDocumentPositionParams, L extends Location = Location>(providers: Observable<ProvideTextDocumentLocationSignature<P, L>[]>, params: P): Observable<L | L[] | null>;
/**
 * Like getLocation, except the returned observable never emits singular values, always either an array or null.
 */
export declare function getLocations<P extends TextDocumentPositionParams = TextDocumentPositionParams, L extends Location = Location>(providers: Observable<ProvideTextDocumentLocationSignature<P, L>[]>, params: P): Observable<L[] | null>;
/**
 * Provides reference results from all extensions.
 *
 * Reference results are always an array or null, unlike results from other location providers (e.g., from
 * textDocument/definition), which can be a single item, an array, or null.
 */
export declare class TextDocumentReferencesProviderRegistry extends TextDocumentLocationProviderRegistry<ReferenceParams> {
    /** Gets reference locations from all extensions. */
    getLocation(params: ReferenceParams): Observable<Location[] | null>;
}
