import { compact, isEqual } from 'lodash'
import { BehaviorSubject, from, isObservable, Observable, of } from 'rxjs'
import { catchError, distinctUntilChanged, map, switchMap } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isPromise, isSubscribable } from '../../util'

/**
 * A status from a status provider with additional information.
 */
export interface WrappedStatus {
    /** The name of the status. */
    name: string

    /** The status. */
    status: sourcegraph.Status
}

/**
 * A service that manages and queries registered status providers
 * ({@link sourcegraph.StatusProvider}).
 */
export interface StatusService {
    /**
     * Observe the statuses provided by all registered providers for a particular scope.
     *
     * The returned observable emits upon subscription and whenever any status changes.
     */
    observeStatuses(scope: Parameters<sourcegraph.StatusProvider['provideStatus']>[0]): Observable<WrappedStatus[]>

    /**
     * Observe the status for a particular provider (by name) and scope.
     *
     * The returned observable emits upon subscription and whenever the status changes.
     */
    observeStatus(
        name: Parameters<typeof sourcegraph.status.registerStatusProvider>[0],
        scope: Parameters<sourcegraph.StatusProvider['provideStatus']>[0]
    ): Observable<WrappedStatus | null>

    /**
     * Register a status provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerStatusProvider: typeof sourcegraph.status.registerStatusProvider
}

/**
 * Creates a new {@link StatusService}.
 */
export function createStatusService(logErrors = true): StatusService {
    interface Registration {
        name: Parameters<typeof sourcegraph.status.registerStatusProvider>[0]
        provider: sourcegraph.StatusProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])

    const provideStatus = (
        registration: Registration,
        args: Parameters<sourcegraph.StatusProvider['provideStatus']>
    ): Observable<WrappedStatus | null> =>
        fromProviderResult(registration.provider.provideStatus(...args), item => item || null).pipe(
            map(status => (status ? { name: registration.name, status } : null)),
            catchError(err => {
                if (logErrors) {
                    console.error(err)
                }
                return [null]
            })
        )

    return {
        observeStatuses: (...args) => {
            return registrations.pipe(
                switchMap(registrations =>
                    combineLatestOrDefault(registrations.map(reg => provideStatus(reg, args))).pipe(
                        map(itemsArrays => compact(itemsArrays)),
                        distinctUntilChanged((a, b) => isEqual(a, b))
                    )
                )
            )
        },
        observeStatus: (name, ...args) => {
            return registrations.pipe(
                switchMap(registrations => {
                    const registration = registrations.find(r => r.name === name)
                    return registration ? provideStatus(registration, args) : of(null)
                })
            )
        },
        registerStatusProvider: (name, provider) => {
            if (registrations.value.some(r => r.name === name)) {
                throw new Error(`a StatusProvider with name ${JSON.stringify(name)} is already registered`)
            }
            const reg: Registration = { name, provider }
            registrations.next([...registrations.value, reg])
            const unregister = () => registrations.next(registrations.value.filter(r => r !== reg))
            return { unsubscribe: unregister }
        },
    }
}

/**
 * Returns an {@link Observable} that represents the same result as a
 * {@link sourcegraph.ProviderResult}, with a mapping.
 *
 * @param result The result returned by the provider
 * @param mapFunc A function to map the result into a type that does not (necessarily) include `|
 * undefined | null`.
 */
function fromProviderResult<T, R>(
    result: sourcegraph.ProviderResult<T>,
    mapFunc: (value: T | undefined | null) => R
): Observable<R> {
    let observable: Observable<R>
    if (result && (isPromise(result) || isObservable<T>(result) || isSubscribable(result))) {
        observable = from(result).pipe(map(mapFunc))
    } else {
        observable = of(mapFunc(result))
    }
    return observable
}
