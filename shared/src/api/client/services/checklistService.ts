import { Observable, BehaviorSubject, from, isObservable, of } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { switchMap, catchError, map, distinctUntilChanged } from 'rxjs/operators'
import { combineLatestOrDefault } from '../../../util/rxjs/combineLatestOrDefault'
import { isEqual, flatten, compact } from 'lodash'
import { isPromise, isSubscribable } from '../../util'

/**
 * A service that manages and queries registered checklist providers
 * ({@link sourcegraph.ChecklistProvider}).
 */
export interface ChecklistService {
    /**
     * Observe the checklist items provided by all registered providers.
     */
    observeChecklistItems(
        scope: Parameters<sourcegraph.ChecklistProvider['provideChecklistItems']>[0]
    ): Observable<sourcegraph.ChecklistItem[]>

    /**
     * Register a checklist provider.
     *
     * @returns An unsubscribable to unregister the provider.
     */
    registerChecklistProvider: typeof sourcegraph.checklist.registerChecklistProvider
}

/**
 * Creates a new {@link ChecklistService}.
 */
export function createChecklistProviderRegistry(logErrors = false): ChecklistService {
    interface Registration {
        type: Parameters<typeof sourcegraph.checklist.registerChecklistProvider>[0]
        provider: sourcegraph.ChecklistProvider
    }
    const registrations = new BehaviorSubject<Registration[]>([])
    return {
        observeChecklistItems: (...args) => {
            return registrations.pipe(
                switchMap(registrations =>
                    combineLatestOrDefault(
                        registrations.map(({ provider }) =>
                            fromProviderResult(provider.provideChecklistItems(...args), items => items || []).pipe(
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
        registerChecklistProvider: (type, provider) => {
            if (registrations.value.some(r => r.type === type)) {
                throw new Error(`a ChecklistProvider of type ${JSON.stringify(type)} is already registered`)
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
