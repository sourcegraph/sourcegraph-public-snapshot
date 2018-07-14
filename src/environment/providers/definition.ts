import { combineLatest, from, Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Definition } from 'vscode-languageserver-types'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentFeatureProviderRegistry } from './textDocument'
import { flattenAndCompact } from './util'

export type ProvideTextDocumentDefinitionSignature = (params: TextDocumentPositionParams) => Promise<Definition | null>

/** Provides definition results from all extensions. */
export class TextDocumentDefinitionProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentDefinitionSignature
> {
    public getDefinition(params: TextDocumentPositionParams): Observable<Definition | null> {
        return getDefinition(of(this.providersSnapshot), params)
    }
}

/**
 * Returns an observable that emits all providers' definition results whenever any of the last-emitted set of
 * providers emits hovers.
 *
 * Most callers should use TextDocumentDefinitionProviderRegistry, which sources definitions from the current set
 * of registered providers (and then completes).
 */
export function getDefinition(
    providers: Observable<ProvideTextDocumentDefinitionSignature[]>,
    params: TextDocumentPositionParams
): Observable<Definition | null> {
    return providers
        .pipe(
            switchMap(providers => {
                if (providers.length === 0) {
                    return [null]
                }
                return combineLatest(
                    providers.map(provider =>
                        from(provider(params)).pipe(
                            catchError(error => {
                                console.error(error)
                                return [null]
                            })
                        )
                    )
                )
            })
        )
        .pipe(map(flattenAndCompact))
}
