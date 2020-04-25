import { Observable, Unsubscribable, BehaviorSubject, of } from 'rxjs'
import { View as ExtensionView } from 'sourcegraph'
import { switchMap, map, distinctUntilChanged } from 'rxjs/operators'

/**
 * A view is a page or partial page.
 */
export interface View extends ExtensionView {}

/**
 * The view service manages views, which are pages or partial pages of content.
 */
export interface ViewService {
    /**
     * Register a view provider. Throws if the `id` is already registered.
     */
    register(id: string, provider: (params: { [key: string]: string }) => Observable<View>): Unsubscribable

    /**
     * Get a view's content. The returned observable emits whenever the content changes. If there is
     * no view with the given {@link id}, it emits `null`.
     */
    get(id: string, params: { [key: string]: string }): Observable<View | null>
}

/**
 * Creates a new {@link ViewService}.
 */
export const createViewService = (): ViewService => {
    type Provider = (params: { [key: string]: string }) => Observable<View>
    const providers = new BehaviorSubject<Map<string, Provider>>(new Map())

    return {
        register: (id, provider) => {
            if (providers.value.has(id)) {
                throw new Error(`view already exists with ID ${id}`)
            }
            providers.value.set(id, provider)
            providers.next(providers.value)
            return {
                unsubscribe: () => {
                    const p = providers.value.get(id)
                    if (p === provider) {
                        // Check equality to ensure we only unsubscribe the exact same provider we
                        // registered, not some other provider that was registered later with the same
                        // ID.
                        providers.value.delete(id)
                        providers.next(providers.value)
                    }
                },
            }
        },
        get: (id, params) =>
            providers.pipe(
                map(providers => providers.get(id)),
                distinctUntilChanged(),
                switchMap(provider => (provider ? provider(params) : of(null)))
            ),
    }
}
