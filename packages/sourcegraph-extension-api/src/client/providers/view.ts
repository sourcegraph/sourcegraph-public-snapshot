import { combineLatest, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { ContributableViewContainer } from '../../protocol'
import * as plain from '../../protocol/plainTypes'
import { Entry, FeatureProviderRegistry } from './registry'

export interface ViewProviderRegistrationOptions {
    id: string
    container: ContributableViewContainer
}

export type ProvideViewSignature = Observable<plain.PanelView>

/** Provides views from all extensions. */
export class ViewProviderRegistry extends FeatureProviderRegistry<
    ViewProviderRegistrationOptions,
    ProvideViewSignature
> {
    /**
     * Returns an observable that emits the specified view whenever it or the set of registered view providers
     * changes. If the provider emits an error, the returned observable also emits an error (and completes).
     */
    public getView(id: string): Observable<plain.PanelView | null> {
        return getView(this.entries, id)
    }

    /**
     * Returns an observable that emits all views whenever the set of registered view providers or their properties
     * change. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    public getViews(
        container: ContributableViewContainer
    ): Observable<(plain.PanelView & ViewProviderRegistrationOptions)[] | null> {
        return getViews(this.entries, container)
    }
}

/**
 * Exported for testing only. Most callers should use {@link ViewProviderRegistry#getView}, which uses the
 * registered view providers.
 *
 * @internal
 */
export function getView(
    entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>,
    id: string
): Observable<(plain.PanelView & ViewProviderRegistrationOptions) | null> {
    return entries.pipe(
        map(entries => entries.find(entry => entry.registrationOptions.id === id)),
        switchMap(entry => (entry ? addRegistrationOptions(entry) : [null]))
    )
}

/**
 * Exported for testing only. Most callers should use {@link ViewProviderRegistry#getViews}, which uses the
 * registered view providers.
 *
 * @internal
 */
export function getViews(
    entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>[]>,
    container: ContributableViewContainer
): Observable<(plain.PanelView & ViewProviderRegistrationOptions)[] | null> {
    return entries.pipe(
        switchMap(
            entries =>
                entries && entries.length > 0
                    ? combineLatest(
                          entries.filter(e => e.registrationOptions.container === container).map(entry =>
                              addRegistrationOptions(entry).pipe(
                                  catchError(err => {
                                      console.error(err)
                                      return [null]
                                  })
                              )
                          )
                      ).pipe(
                          map(entries =>
                              entries.filter(
                                  (result): result is plain.PanelView & ViewProviderRegistrationOptions =>
                                      result !== null
                              )
                          )
                      )
                    : [null]
        )
    )
}

function addRegistrationOptions(
    entry: Entry<ViewProviderRegistrationOptions, Observable<plain.PanelView>>
): Observable<plain.PanelView & ViewProviderRegistrationOptions> {
    return entry.provider.pipe(map(view => ({ ...view, ...entry.registrationOptions })))
}
