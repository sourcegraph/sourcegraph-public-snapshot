import { Range, Selection } from '@sourcegraph/extension-api-classes'
import { isEqual } from 'lodash'
import { combineLatest, from, Observable, of } from 'rxjs'
import { catchError, defaultIfEmpty, distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators'
import { Action, CodeActionContext } from 'sourcegraph'
import { ErrorLike } from '../../../util/errors'
import { TextDocumentIdentifier } from '../types/textDocument'
import { DocumentFeatureProviderRegistry } from './registry'
import { flattenAndCompact } from './util'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'

export interface CodeActionsParams {
    textDocument: TextDocumentIdentifier
    range: Range | Selection
    context: CodeActionContext
}

const codeActionError = Symbol('isCodeActionError')

export interface CodeActionError extends ErrorLike {
    [codeActionError]: true
    params: CodeActionsParams
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function isCodeActionError(val: any): val is CodeActionError {
    return val[codeActionError] === true
}

export type ProvideCodeActionsSignature = (params: CodeActionsParams) => Observable<Action[] | null | undefined>

/** Provides code actions from all extensions. */
export class CodeActionProviderRegistry extends DocumentFeatureProviderRegistry<ProvideCodeActionsSignature> {
    /**
     * Returns an observable that emits all providers' results whenever any of the last-emitted set
     * of providers emits code actions. If any provider emits an error, the error is logged and the
     * provider result is omitted from the emission of the observable (the observable does not emit
     * the error).
     */
    public getCodeActions(params: CodeActionsParams): Observable<(Action | CodeActionError)[] | null> {
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
): Observable<(Action | CodeActionError)[] | null> {
    return providers.pipe(
        filter(providers => providers.length > 0),
        switchMap(providers =>
            combineLatestOrDefault(
                providers.map(provider =>
                    from(
                        provider(params).pipe(
                            catchError(err => {
                                if (logErrors) {
                                    console.error(err)
                                }
                                return of<CodeActionError[]>([
                                    { [codeActionError]: true, message: err.message, params },
                                ])
                            })
                        )
                    )
                )
            ).pipe(
                map(a => flattenAndCompact<Action | CodeActionError>(a)),
                defaultIfEmpty<(Action | CodeActionError)[] | null>(null),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        )
    )
}
