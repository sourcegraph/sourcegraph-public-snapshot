import { Hover } from '@sourcegraph/extension-api-types'
import { from, Observable } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
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
): Observable<HoverMerged | null> {
    return providers.pipe(
        switchMap(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    from(
                        provider(params).pipe(
                            catchError(err => {
                                if (logErrors) {
                                    console.error(err)
                                }
                                return [null]
                            })
                        )
                    )
                )
            ).pipe(
                map(HoverMerged.from),
                defaultIfEmpty(null as HoverMerged | null),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
}
