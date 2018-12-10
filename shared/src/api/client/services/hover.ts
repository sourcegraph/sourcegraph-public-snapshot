import { Hover } from '@sourcegraph/extension-api-types'
import { combineLatest, from, Observable } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, filter, map, startWith, switchMap } from 'rxjs/operators'
import { HoverMerged } from '../../client/types/hover'
import { TextDocumentPositionParams } from '../../protocol'
import { isEqual } from '../../util'
import { DocumentFeatureProviderRegistry } from './registry'

export type ProvideTextDocumentHoverSignature = (
    params: TextDocumentPositionParams
) => Observable<Hover | null | undefined>

/** Provides hovers from all extensions. */
export class TextDocumentHoverProviderRegistry extends DocumentFeatureProviderRegistry<
    ProvideTextDocumentHoverSignature
> {
    /**
     * Returns an observable that emits all providers' hovers whenever any of the last-emitted set of providers emits
     * hovers. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    public getHover(params: TextDocumentPositionParams): Observable<HoverMerged | null> {
        return getHover(this.providersForDocument(params.textDocument), params)
    }
}

const INITIAL = Symbol('INITIAL')

/**
 * Returns an observable that emits all providers' hovers whenever any of the last-emitted set of providers emits
 * hovers. If any provider emits an error, the error is logged and the provider is omitted from the emission of
 * the observable (the observable does not emit the error).
 *
 * Most callers should use TextDocumentHoverProviderRegistry's getHover method, which uses the registered hover
 * providers.
 */
export function getHover(
    providers: Observable<ProvideTextDocumentHoverSignature[]>,
    params: TextDocumentPositionParams
): Observable<HoverMerged | null> {
    return providers.pipe(
        switchMap(providers => {
            if (providers.length === 0) {
                return [null]
            }
            return combineLatest(
                providers.map(provider =>
                    from(
                        provider(params).pipe(
                            // combineLatest waits to emit until all observables have emitted. Make all
                            // observables emit immediately to avoid waiting for the slowest observable.
                            startWith(INITIAL),

                            catchError(err => {
                                console.error(err)
                                return [null]
                            })
                        )
                    )
                )
            ).pipe(
                filter(results => results === null || !results.every(result => result === INITIAL)),
                map(
                    results =>
                        results && results.filter((result): result is Hover | null | undefined => result !== INITIAL)
                ),
                map(HoverMerged.from),
                defaultIfEmpty(null as HoverMerged | null),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        })
    )
}
