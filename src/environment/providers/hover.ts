import { combineLatest, from, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Hover } from 'vscode-languageserver-types'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { HoverMerged } from '../../types/hover'
import { TextDocumentFeatureProviderRegistry } from './textDocument'

export type ProvideTextDocumentHoverSignature = (params: TextDocumentPositionParams) => Promise<Hover | null>

/** Provides hovers from all extensions. */
export class TextDocumentHoverProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentHoverSignature
> {
    public getHover(params: TextDocumentPositionParams): Observable<HoverMerged | null> {
        return this.providersSnapshot
            .pipe(
                switchMap(providers =>
                    combineLatest(
                        providers.map(provider =>
                            from(provider(params)).pipe(
                                catchError(error => {
                                    console.error(error)
                                    return [null]
                                })
                            )
                        )
                    )
                )
            )
            .pipe(map(HoverMerged.from))
    }
}
