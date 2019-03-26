import { Location } from '@sourcegraph/extension-api-types'
import React from 'react'
import { Observable } from 'rxjs'
import { catchError, map, switchMap, tap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ErrorLike } from '../../../util/errors'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { ContributableViewContainer } from '../../protocol'
import { Entry, FeatureProviderRegistry } from './registry'

export interface ViewProviderRegistrationOptions {
    id: string
    container: ContributableViewContainer
}

export interface PanelViewWithComponent extends Pick<sourcegraph.PanelView, 'title' | 'content' | 'priority'> {
    locationsOrCustom: { locations: Observable<Observable<Location[] | null>> } | { custom: React.ReactFragment }
}

export type ProvideViewSignature = Observable<PanelViewWithComponent | null>

/** Provides views from all extensions. */
export class ViewProviderRegistry extends FeatureProviderRegistry<
    ViewProviderRegistrationOptions,
    ProvideViewSignature
> {
    /**
     * Returns an observable that emits the specified view whenever it or the set of registered view providers
     * changes. If the provider emits an error, the returned observable also emits an error (and completes).
     */
    public getView(id: string): Observable<PanelViewWithComponent | null> {
        return getView(this.entries, id)
    }

    /**
     * Returns an observable that emits all views whenever the set of registered view providers or their properties
     * change. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    public getViews(
        container: ContributableViewContainer
    ): Observable<(PanelViewWithComponent & ViewProviderRegistrationOptions)[]> {
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
    entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent | null>>[]>,
    id: string
): Observable<(PanelViewWithComponent & ViewProviderRegistrationOptions) | null> {
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
    entries: Observable<Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent | null>>[]>,
    container: ContributableViewContainer,
    logErrors = true
): Observable<(PanelViewWithComponent & ViewProviderRegistrationOptions)[]> {
    return entries.pipe(
        switchMap(entries =>
            combineLatestOrDefault(
                entries
                    .filter(e => e.registrationOptions.container === container)
                    .map(entry =>
                        addRegistrationOptions(entry).pipe(
                            catchError(err => {
                                if (logErrors) {
                                    console.error(err)
                                }
                                return [null]
                            })
                        )
                    )
            ).pipe(
                map(entries =>
                    entries.filter(
                        (result): result is PanelViewWithComponent & ViewProviderRegistrationOptions => result !== null
                    )
                )
            )
        )
    )
}

function addRegistrationOptions(
    entry: Entry<ViewProviderRegistrationOptions, Observable<PanelViewWithComponent | null>>
): Observable<(PanelViewWithComponent & ViewProviderRegistrationOptions) | null> {
    return entry.provider.pipe(map(view => view && { ...view, ...entry.registrationOptions }))
}
