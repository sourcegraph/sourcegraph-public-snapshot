import { Observable } from 'rxjs';
import { ContributableViewContainer } from '../../protocol';
import * as plain from '../../protocol/plainTypes';
import { Entry, FeatureProviderRegistry } from './registry';
export interface ViewProviderRegistrationOptions {
    id: string;
    container: ContributableViewContainer;
}
export declare type ProvideViewSignature = Observable<plain.PanelView>;
/** Provides views from all extensions. */
export declare class ViewProviderRegistry extends FeatureProviderRegistry<ViewProviderRegistrationOptions, ProvideViewSignature> {
    /**
     * Returns an observable that emits the specified view whenever it or the set of registered view providers
     * changes. If the provider emits an error, the returned observable also emits an error (and completes).
     */
    getView(id: string): Observable<plain.PanelView | null>;
    /**
     * Returns an observable that emits all views whenever the set of registered view providers or their properties
     * change. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    getViews(container: ContributableViewContainer): Observable<(plain.PanelView & ViewProviderRegistrationOptions)[] | null>;
}
/**
 * Exported for testing only. Most callers should use {@link ViewProviderRegistry#getView}, which uses the
 * registered view providers.
 *
 * @internal
 */
export declare function getView(entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>, id: string): Observable<(plain.PanelView & ViewProviderRegistrationOptions) | null>;
/**
 * Exported for testing only. Most callers should use {@link ViewProviderRegistry#getViews}, which uses the
 * registered view providers.
 *
 * @internal
 */
export declare function getViews(entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>, container: ContributableViewContainer): Observable<(plain.PanelView & ViewProviderRegistrationOptions)[] | null>;
