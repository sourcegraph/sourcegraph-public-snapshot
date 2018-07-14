import { combineLatest, Observable } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentDecoration, TextDocumentDecorationsParams } from '../../protocol/decorations'
import { compact, flatten } from '../../util'
import { TextDocumentFeatureProviderRegistry } from './textDocument'

export type ProvideTextDocumentDecorationsSignature = (
    params: TextDocumentDecorationsParams
) => Observable<TextDocumentDecoration[] | null>

/** Provides text document decorations from all extensions. */
export class TextDocumentDecorationsProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentDecorationsSignature
> {
    public getDecorations(params: TextDocumentDecorationsParams): Observable<TextDocumentDecoration[] | null> {
        return this.providers
            .pipe(
                switchMap(providers =>
                    combineLatest(
                        providers.map(provider =>
                            provider(params).pipe(
                                map(results => (results === null ? [] : compact(results))),
                                catchError(error => {
                                    console.error(error)
                                    return [[]]
                                })
                            )
                        )
                    )
                )
            )
            .pipe(map(results => flatten(results)))
    }
}
