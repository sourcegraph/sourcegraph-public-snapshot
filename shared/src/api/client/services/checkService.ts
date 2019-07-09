import { isEqual } from 'lodash'
import { BehaviorSubject, from, Observable, of, Unsubscribable, Subscribable } from 'rxjs'
import { distinctUntilChanged, map, switchMap, catchError } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { ErrorLike, asError } from '../../../util/errors'

export interface CheckID {
    type: Parameters<typeof sourcegraph.checks.registerCheckProvider>[0]
    id: sourcegraph.CheckContext<any>['id']
}

export type CheckProviderOrError = CheckID &
    ({ provider: sourcegraph.CheckProvider; error?: undefined } | { provider?: undefined; error: ErrorLike })

/**
 * A service that manages and queries registered check providers
 * ({@link sourcegraph.CheckProvider}).
 */
export interface CheckService {
    /**
     * Observe all checks provided by all registered providers for a particular scope.
     *
     * The returned observable emits upon subscription and whenever any check changes.
     */
    observeChecks(scope: sourcegraph.CheckContext<any>['scope']): Observable<CheckProviderOrError[]>

    /**
     * Observe a single check.
     *
     * The returned observable emits upon subscription and whenever the check changes.
     *
     * //TODO!(sqs): make this handle errors in invoking the providerFactory
     */
    observeCheck(
        scope: sourcegraph.CheckContext<any>['scope'],
        id: CheckID
    ): Observable<sourcegraph.CheckProvider | null>

    /**
     * Register a check provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerCheckProvider(
        type: CheckID['type'],
        providerFactory: (context: sourcegraph.CheckContext<any>) => sourcegraph.CheckProvider
    ): Unsubscribable
}

// TODO!(sqs)
const DUMMY_CONTEXT: Pick<sourcegraph.CheckContext<{}>, 'id'> = {
    id: 'DUMMY',
}

/**
 * Creates a new {@link CheckService}.
 */
export function createCheckService(): CheckService {
    interface Registration {
        type: CheckID['type']
        providerFactory: (context: sourcegraph.CheckContext<any>) => sourcegraph.CheckProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])

    const providerFromFactory = (
        { type, providerFactory }: Pick<Registration, 'type' | 'providerFactory'>,
        context: sourcegraph.CheckContext<any>
    ): CheckProviderOrError => {
        try {
            const provider = providerFactory(context)
            return { type, id: context.id, provider }
        } catch (err) {
            return { type, id: context.id, error: asError(err) }
        }
    }

    return {
        observeChecks: scope => {
            return registrations
                .pipe(
                    map(registrations =>
                        registrations.map(registration =>
                            providerFromFactory(registration, { ...DUMMY_CONTEXT, scope })
                        )
                    )
                )
                .pipe(distinctUntilChanged((a, b) => isEqual(a, b)))
        },
        observeCheck: (scope, { type, id }) => {
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

export type CheckInformationOrError = CheckID &
    ({ information: sourcegraph.CheckInformation; error?: undefined } | { information?: undefined; error: ErrorLike })

export function observeChecksInformation(
    service: Pick<CheckService, 'observeChecks'>,
    scope: sourcegraph.CheckContext<any>['scope']
): Subscribable<CheckInformationOrError[]> {
    return service.observeChecks(scope).pipe(
        switchMap(checks =>
            combineLatestOrDefault(
                checks.map<Subscribable<CheckInformationOrError>>(({ type, id, ...other }) =>
                    other.provider
                        ? from(other.provider.information).pipe(
                              map(information => ({
                                  type,
                                  id,
                                  information,
                              })),
                              catchError(err => {
                                  return of({ type, id, error: asError(err) })
                              })
                          )
                        : of({ type, id, error: other.error })
                )
            ).pipe(distinctUntilChanged((a, b) => isEqual(a, b)))
        )
    )
}
