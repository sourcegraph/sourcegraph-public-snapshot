import { Observable, BehaviorSubject, from, isObservable, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { switchMap, catchError, map, distinctUntilChanged } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isEqual, compact } from 'lodash'
import { isPromise, isSubscribable } from '../../util'

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
    observeStatuses(scope: Parameters<sourcegraph.StatusProvider['provideStatus']>[0]): Observable<sourcegraph.Status[]>

    /**
     * Observe the status for a particular provider (by type) and scope.
     *
     * The returned observable emits upon subscription and whenever the status changes.
     */
    observeStatus(
        type: Parameters<typeof sourcegraph.status.registerStatusProvider>[0],
        scope: Parameters<sourcegraph.StatusProvider['provideStatus']>[0]
    ): Observable<sourcegraph.Status | null>

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
    const provideStatus = (
        provider: sourcegraph.StatusProvider,
        args: Parameters<sourcegraph.StatusProvider['provideStatus']>
    ): Observable<sourcegraph.Status | null> =>
        fromProviderResult(provider.provideStatus(...args), item => item || null).pipe(
            catchError(err => {
                if (logErrors) {
                    console.error(err)
                }
                return [null]
            })
        )

    interface Registration {
        type: Parameters<typeof sourcegraph.status.registerStatusProvider>[0]
        provider: sourcegraph.StatusProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])
    return {
        observeStatuses: (...args) => {
            return registrations.pipe(
                switchMap(registrations =>
                    combineLatestOrDefault(registrations.map(({ provider }) => provideStatus(provider, args))).pipe(
                        map(itemsArrays => compact(itemsArrays)),
                        distinctUntilChanged((a, b) => isEqual(a, b))
                    )
                )
            )
        },
        observeStatus: (type, ...args) => {
            return registrations.pipe(
                switchMap(registrations => {
                    const reg = registrations.find(r => r.type === type)
                    return reg ? provideStatus(reg.provider, args) : of(null)
                })
            )
        },
        registerStatusProvider: (type, provider) => {
            if (registrations.value.some(r => r.type === type)) {
                throw new Error(`a StatusProvider of type ${JSON.stringify(type)} is already registered`)
            }
            const reg: Registration = { type, provider }
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
