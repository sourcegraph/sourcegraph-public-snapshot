import { Observable, BehaviorSubject, from, combineLatest } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { switchMap, catchError, map, distinctUntilChanged } from 'rxjs/operators'
import { isEqual, flatten, compact } from 'lodash'

export interface DiagnosticWithType extends sourcegraph.Diagnostic {
    /**
     * The type of the diagnostic provider that produced this diagnostic.
     */
    type: string
}

/**
 * A service that manages and queries registered diagnostic providers
 * ({@link sourcegraph.DiagnosticProvider}).
 */
export interface DiagnosticService {
    /**
     * Observe the diagnostics provided by registered providers or by a specific provider.
     *
     * @param type Only observe diagnostics from the provider registered with this type. If
     * undefined, diagnostics from all providers are observed.
     */
    observeDiagnostics(
        scope: Parameters<sourcegraph.DiagnosticProvider['provideDiagnostics']>[0],
        context: Parameters<sourcegraph.DiagnosticProvider['provideDiagnostics']>[1],
        type?: Parameters<typeof sourcegraph.workspace.registerDiagnosticProvider>[0]
    ): Observable<DiagnosticWithType[]>

    /**
     * Register a diagnostic provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerDiagnosticProvider: typeof sourcegraph.workspace.registerDiagnosticProvider
}

/**
 * Creates a new {@link DiagnosticService}.
 */
export function createDiagnosticService(logErrors = true): DiagnosticService {
    interface Registration {
        type: Parameters<typeof sourcegraph.workspace.registerDiagnosticProvider>[0]
        provider: sourcegraph.DiagnosticProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])
    return {
        observeDiagnostics: (scope, context, type) => {
            return registrations.pipe(
                switchMap(registrations =>
                    // Use combineLatest not combineLatestOrDefault because we don't want to falsely
                    // report "0 diagnostics" if a provider has not been registered yet.
                    combineLatest(
                        (type === undefined ? registrations : registrations.filter(r => r.type === type)).map(
                            ({ type, provider }) =>
                                from(provider.provideDiagnostics(scope, context)).pipe(
                                    map(diagnostics => diagnostics.map(diagnostic => ({ ...diagnostic, type }))),
                                    catchError(err => {
                                        if (logErrors) {
                                            console.error(err)
                                        }
                                        return [null]
                                    })
                                )
                        )
                    ).pipe(
                        map(itemsArrays => flatten(compact(itemsArrays))),
                        distinctUntilChanged((a, b) => isEqual(a, b))
                    )
                )
            )
        },
        registerDiagnosticProvider: (type, provider) => {
            if (registrations.value.some(r => r.type === type)) {
                throw new Error(`a DiagnosticProvider of type ${JSON.stringify(type)} is already registered`)
            }
            const reg: Registration = { type, provider }
            registrations.next([...registrations.value, reg])
            const unregister = () => registrations.next(registrations.value.filter(r => r !== reg))
            return { unsubscribe: unregister }
        },
    }
}
