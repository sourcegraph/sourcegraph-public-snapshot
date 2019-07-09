import { isEqual, compact } from 'lodash'
import { BehaviorSubject, from, Observable, of, Unsubscribable, Subscribable } from 'rxjs'
import { distinctUntilChanged, map, switchMap, catchError } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isDefined } from '../../../util/types'

export interface CheckInformationWithID extends sourcegraph.CheckInformation {
    type: string
    id: string
}

/**
 * A service that manages and queries registered check providers
 * ({@link sourcegraph.CheckProvider}).
 */
export interface CheckService {
    /**
     * Observe information for all checks provided by all registered providers for a particular scope.
     *
     * The returned observable emits upon subscription and whenever any check changes.
     */
    observeChecksInformation(scope: sourcegraph.CheckContext<any>['scope']): Observable<CheckInformationWithID[]>

    /**
     * Observe the check for a particular provider (by type).
     *
     * The returned observable emits upon subscription and whenever the check changes.
     */
    observeCheck(
        type: Parameters<typeof sourcegraph.checks.registerCheckProvider>[0],
        scope: sourcegraph.CheckContext<any>['scope'],
        id: sourcegraph.CheckContext<any>['id']
    ): Observable<sourcegraph.CheckProvider | null>

    /**
     * Register a check provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerCheckProvider(
        type: Parameters<typeof sourcegraph.checks.registerCheckProvider>[0],
        providerFactory: (context: sourcegraph.CheckContext<any>) => sourcegraph.CheckProvider
    ): Unsubscribable
}

// TODO!(sqs)
const DUMMY_CONTEXT: Pick<sourcegraph.CheckContext<{}>, 'id' | 'settings'> = {
    id: 'DUMMY',
    settings: of({}),
}

/**
 * Creates a new {@link CheckService}.
 */
export function createCheckService(logErrors = true): CheckService {
    interface Registration {
        type: Parameters<typeof sourcegraph.checks.registerCheckProvider>[0]
        providerFactory: (context: sourcegraph.CheckContext<any>) => sourcegraph.CheckProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])

    return {
        observeChecksInformation: scope => {
            return registrations.pipe(
                switchMap(registrations =>
                    combineLatestOrDefault(
                        registrations.map(registration =>
                            from(registration.providerFactory({ ...DUMMY_CONTEXT, scope }).information).pipe(
                                map(
                                    info =>
                                        ({
                                            ...info,
                                            type: registration.type,
                                            id: DUMMY_CONTEXT.id,
                                        } as CheckInformationWithID)
                                ),
                                catchError(err => {
                                    if (logErrors) {
                                        console.error(err)
                                    }
                                    return of(null)
                                })
                            )
                        )
                    ).pipe(
                        map(results => compact(results)),
                        distinctUntilChanged((a, b) => isEqual(a, b))
                    )
                )
            )
        },
        observeCheck: (type, scope, id) => {
            return registrations.pipe(
                map(registrations => {
                    const registration = registrations.find(r => r.type === type)
                    return registration ? registration.providerFactory({ ...DUMMY_CONTEXT, scope, id }) : null
                })
            )
        },
        registerCheckProvider: (type, providerFactory) => {
            if (registrations.value.some(r => r.type === type)) {
                throw new Error(`a CheckProvider with type ${JSON.stringify(type)} is already registered`)
            }
            const registration: Registration = { type, providerFactory }
            registrations.next([...registrations.value, registration])
            const unregister = () => registrations.next(registrations.value.filter(r => r !== registration))
            return { unsubscribe: unregister }
        },
    }
}
