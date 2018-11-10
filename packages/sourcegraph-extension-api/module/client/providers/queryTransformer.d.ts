import { Observable } from 'rxjs';
import { FeatureProviderRegistry } from './registry';
export declare type TransformQuerySignature = (query: string) => Observable<string>;
/** Transforms search queries using registered query transformers from extensions. */
export declare class QueryTransformerRegistry extends FeatureProviderRegistry<{}, TransformQuerySignature> {
    transformQuery(query: string): Observable<string>;
}
/**
 * Returns an observable that emits a query transformed by all providers whenever any of the last-emitted set of providers emits
 * a query.
 *
 * Most callers should use QueryTransformerRegistry's transformQuery method, which uses the registered query transformers
 *
 */
export declare function transformQuery(providers: Observable<TransformQuerySignature[]>, query: string): Observable<string>;
