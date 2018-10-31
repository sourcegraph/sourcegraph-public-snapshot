import { combineLatest, from, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { HoverMerged } from '../../client/types/hover'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { Hover } from '../../protocol/plainTypes'
import { FeatureProviderRegistry } from './registry'

export type ProvideTextDocumentHoverSignature = (
    params: TextDocumentPositionParams
) => Observable<Hover | null | undefined>

/** Provides hovers from all extensions. */
export class TextDocumentHoverProviderRegistry extends FeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentHoverSignature
> {
    /**
     * Returns an observable that emits all providers' hovers whenever any of the last-emitted set of providers emits
     * hovers. If any provider emits an error, the error is logged and the provider is omitted from the emission of
     * the observable (the observable does not emit the error).
     */
    public getHover(params: TextDocumentPositionParams): Observable<HoverMerged | null> {
        return getHover(this.providers, params)
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
    params: TextDocumentPositionParams
): Observable<HoverMerged | null> {
    return providers
        .pipe(
            switchMap(providers => {
                if (providers.length === 0) {
                    return [[null]]
                }
                return combineLatest(
                    providers.map(provider =>
                        from(
                            provider(params).pipe(
                                catchError(err => {
                                    console.error(err)
                                    return [null]
                                })
                            )
                        )
                    )
                )
            })
        )
        .pipe(map(HoverMerged.from))
}
