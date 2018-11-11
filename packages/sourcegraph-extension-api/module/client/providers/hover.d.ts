import { Observable } from 'rxjs';
import { HoverMerged } from '../../client/types/hover';
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol';
import { Hover } from '../../protocol/plainTypes';
import { FeatureProviderRegistry } from './registry';
export declare type ProvideTextDocumentHoverSignature = (params: TextDocumentPositionParams) => Observable<Hover | null | undefined>;
/** Provides hovers from all extensions. */
export declare class TextDocumentHoverProviderRegistry extends FeatureProviderRegistry<TextDocumentRegistrationOptions, ProvideTextDocumentHoverSignature> {
    /**
     * Returns an observable that emits all providers' hovers whenever any of the last-emitted set of providers emits
     * hovers. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    getHover(params: TextDocumentPositionParams): Observable<HoverMerged | null>;
}
/**
 * Returns an observable that emits all providers' hovers whenever any of the last-emitted set of providers emits
 * hovers. If any provider emits an error, the error is logged and the provider is omitted from the emission of
 * the observable (the observable does not emit the error).
 *
 * Most callers should use TextDocumentHoverProviderRegistry's getHover method, which uses the registered hover
 * providers.
 */
export declare function getHover(providers: Observable<ProvideTextDocumentHoverSignature[]>, params: TextDocumentPositionParams): Observable<HoverMerged | null>;
