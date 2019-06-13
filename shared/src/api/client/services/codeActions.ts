import { Range, Selection } from '@sourcegraph/extension-api-types'
import { isEqual } from 'lodash'
import { from, Observable } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import { CodeAction, CodeActionContext } from 'sourcegraph'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { TextDocumentIdentifier } from '../types/textDocument'
import { DocumentFeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'

export interface CodeActionsParams {
    textDocument: TextDocumentIdentifier
    range: Range | Selection
    context: CodeActionContext
}

export type ProvideCodeActionsSignature = (params: CodeActionsParams) => Observable<CodeAction[] | null | undefined>

/** Provides code actions from all extensions. */
export class CodeActionProviderRegistry extends DocumentFeatureProviderRegistry<ProvideCodeActionsSignature> {
    /**
     * Returns an observable that emits all providers' results whenever any of the last-emitted set
     * of providers emits code actions. If any provider emits an error, the error is logged and the
     * provider result is omitted from the emission of the observable (the observable does not emit
     * the error).
     */
    public getCodeActions(params: CodeActionsParams): Observable<CodeAction[] | null> {
        return getCodeActions(this.providersForDocument(params.textDocument), params)
    }
}

/**
 * Returns an observable that emits all providers' completion items whenever any of the last-emitted
 * set of providers emits completion items. If any provider emits an error, the error is logged and
 * the provider is omitted from the emission of the observable (the observable does not emit the
 * error).
 *
 * Most callers should use {@link CodeActionsProviderRegistry#getCodeActions}, which uses
 * the registered providers.
 */
export function getCodeActions(
    providers: Observable<ProvideCodeActionsSignature[]>,
    params: CodeActionsParams,
    logErrors = true
): Observable<CodeAction[] | null> {
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
                map(flattenAndCompact),
                defaultIfEmpty<CodeAction[] | null>(null),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
}
