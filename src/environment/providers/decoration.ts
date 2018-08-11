import { combineLatest, Observable } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'
import { TextDocumentIdentifier } from 'vscode-languageserver-types'
import { TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentDecoration } from '../../protocol/decoration'
import { TextDocumentFeatureProviderRegistry } from './textDocument'
import { flattenAndCompact } from './util'

export type ProvideTextDocumentDecorationSignature = (
    textDocument: TextDocumentIdentifier
) => Observable<TextDocumentDecoration[] | null>

/** Provides text document decorations from all extensions. */
export class TextDocumentDecorationProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentDecorationSignature
> {
    public getDecorations(params: TextDocumentIdentifier): Observable<TextDocumentDecoration[] | null> {
        return getDecorations(this.providers, params)
    }
}

/**
 * Returns an observable that emits all decorations whenever any of the last-emitted set of providers emits
 * decorations.
 *
 * Most callers should use TextDocumentDecorationProviderRegistry, which uses the registered decoration providers.
 */
export function getDecorations(
    providers: Observable<ProvideTextDocumentDecorationSignature[]>,
    params: TextDocumentIdentifier
): Observable<TextDocumentDecoration[] | null> {
    return providers
        .pipe(
            switchMap(providers => {
                if (providers.length === 0) {
                    return [null]
                }
                return combineLatest(providers.map(provider => provider(params)))
            })
        )
        .pipe(map(flattenAndCompact))
}
