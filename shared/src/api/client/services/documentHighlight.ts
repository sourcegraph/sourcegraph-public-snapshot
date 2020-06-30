import { isEqual } from 'lodash'
import { from, Observable } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { DocumentHighlight } from 'sourcegraph'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isDefined } from '../../../util/types'
import { TextDocumentPositionParams } from '../../protocol'
import { DocumentFeatureProviderRegistry } from './registry'

export type ProvideDocumentHighlightSignature = (
    params: TextDocumentPositionParams
) => Observable<DocumentHighlight[] | null | undefined>

/** Provides document highlights from all extensions. */
export class DocumentHighlightProviderRegistry extends DocumentFeatureProviderRegistry<
    ProvideDocumentHighlightSignature
> {
    /**
     * Returns an observable that emits all providers' results whenever any of the last-emitted set
     * of providers emits document highlights. If any provider emits an error, the error is logged and the
     * provider result is omitted from the emission of the observable (the observable does not emit
     * the error).
     */
    public getDocumentHighlights(parameters: TextDocumentPositionParams): Observable<DocumentHighlight[] | null> {
        return getDocumentHighlights(this.providersForDocument(parameters.textDocument), parameters)
    }
}

/**
 * Returns an observable that emits all providers' document highlights whenever any of the last-emitted
 * set of providers emits document hovers. If any provider emits an error, the error is logged and
 * the provider is omitted from the emission of the observable (the observable does not emit the
 * error).
 *
 * Most callers should use {@link DocumentHighlightProviderRegistry#getDocumentHighlights}, which uses
 * the registered providers.
 */
export function getDocumentHighlights(
    providers: Observable<ProvideDocumentHighlightSignature[]>,
    parameters: TextDocumentPositionParams,
    logErrors = true
): Observable<DocumentHighlight[] | null> {
    return providers.pipe(
        switchMap(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    from(
                        provider(parameters).pipe(
                            catchError(error => {
                                if (logErrors) {
                                    console.error(error)
                                }
                                return [null]
                            })
                        )
                    )
                )
            ).pipe(
                map(mergeDocumentHighlights),
                defaultIfEmpty<DocumentHighlight[] | null>(null),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
}

function mergeDocumentHighlights(values: (DocumentHighlight[] | null | undefined)[]): DocumentHighlight[] | null {
    const items = values.filter(isDefined).flatMap(items => items)

    return items.length > 0 ? items : null
}
