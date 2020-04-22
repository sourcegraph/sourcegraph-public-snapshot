import { Hover } from '@sourcegraph/extension-api-types'
import { isEqual } from 'lodash'
import { Observable, concat } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { fromHoverMerged, HoverMerged } from '../types/hover'
import { TextDocumentPositionParams } from '../../protocol'
import { DocumentFeatureProviderRegistry } from './registry'
import { isNot, isExactly } from '../../../util/types'
import { MaybeLoadingResult, LOADING } from '@sourcegraph/codeintellify'

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
    public getHover(params: TextDocumentPositionParams): Observable<MaybeLoadingResult<HoverMerged | null>> {
        return getHover(this.providersForDocument(params.textDocument), params)
    }
}

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
    params: TextDocumentPositionParams,
    logErrors = true
): Observable<MaybeLoadingResult<HoverMerged | null>> {
    return providers.pipe(
        switchMap(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    concat(
                        [LOADING],
                        provider(params).pipe(
                            defaultIfEmpty<typeof LOADING | Hover | null | undefined>(null),
                            catchError(err => {
                                if (logErrors) {
                                    console.error('Hover provider errored:', err)
                                }
                                return [null]
                            })
                        )
                    )
                )
            ).pipe(
                defaultIfEmpty<(typeof LOADING | Hover | null | undefined)[]>([]),
                map(hoversFromProviders => ({
                    isLoading: hoversFromProviders.some(hover => hover === LOADING),
                    result: fromHoverMerged(hoversFromProviders.filter(isNot(isExactly(LOADING)))),
                })),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
}
