import { combineLatest, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentDecoration, TextDocumentDecorationParams } from '../../protocol/decoration'
import { compact, flatten } from '../../util'
import { TextDocumentFeatureProviderRegistry } from './textDocument'

export type ProvideTextDocumentDecorationSignature = (
    params: TextDocumentDecorationParams
) => Observable<TextDocumentDecoration[] | null>

/** Provides text document decorations from all extensions. */
export class TextDocumentDecorationProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentDecorationSignature
> {
    public getDecorations(params: TextDocumentDecorationParams): Observable<TextDocumentDecoration[] | null> {
        return getDecorations(this.providers, params)
    }
}

/**
 * Returns an observable that emits all decorations whenever any of the last-emitted set of providers emits
 * decorations.
 *
 * Most callers should use TextDocumentDecorationProviderRegistry, which sources decorations from the set of
 * registered providers.
 */
export function getDecorations(
    providers: Observable<ProvideTextDocumentDecorationSignature[]>,
    params: TextDocumentDecorationParams
): Observable<TextDocumentDecoration[] | null> {
    return providers
        .pipe(
            switchMap(providers => {
                if (providers.length === 0) {
                    return [null]
                }
                return combineLatest(
                    providers.map(provider =>
                        provider(params).pipe(
                            map(results => (results === null ? [] : compact(results))),
                            catchError(error => {
                                console.error(error)
                                return [null]
                            })
                        )
                    )
                )
            })
        )
        .pipe(
            map(results => {
                if (results === null) {
                    return null
                }
                const merged = flatten(compact(results))
                return merged.length === 0 ? null : merged
            })
        )
}
