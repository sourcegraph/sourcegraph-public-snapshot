import { combineLatest, from, Observable, of } from 'rxjs'
import { catchError, map, switchMap } from 'rxjs/operators'
import { Definition } from 'vscode-languageserver-types'
import { TextDocumentPositionParams, TextDocumentRegistrationOptions } from '../../protocol'
import { TextDocumentFeatureProviderRegistry } from './textDocument'
import { flattenAndCompact } from './util'

/**
 * Function signature for retrieving related locations given a location (e.g., definition, implementation, and type
 * definition).
 */
export type ProvideTextDocumentLocationSignature = (params: TextDocumentPositionParams) => Promise<Definition | null>

/** Provides location results from all extensions for definition, implementation, and type definition requests. */
export class TextDocumentLocationProviderRegistry extends TextDocumentFeatureProviderRegistry<
    TextDocumentRegistrationOptions,
    ProvideTextDocumentLocationSignature
> {
    public getLocation(params: TextDocumentPositionParams): Observable<Definition | null> {
        return getLocation(of(this.providersSnapshot), params)
    }
}

/**
 * Returns an observable that emits all providers' location results whenever any of the last-emitted set of
 * providers emits hovers.
 *
 * Most callers should use the TextDocumentLocationProviderRegistry class, which sources results from the current
 * set of registered providers (and then completes).
 */
export function getLocation(
    providers: Observable<ProvideTextDocumentLocationSignature[]>,
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
