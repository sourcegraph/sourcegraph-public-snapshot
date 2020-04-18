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
    interface ProviderEntry {
        id: string
        provider: (params: { [key: string]: string }) => Observable<View>
    }
    const providers = new BehaviorSubject<ProviderEntry[]>([])

    return {
        register: (id, provider) => {
            if (providers.value.some(e => e.id === id)) {
                throw new Error(`view already exists with ID ${id}`)
            }
            const entry: ProviderEntry = { id, provider }
            providers.next([...providers.value, entry])
            return { unsubscribe: () => providers.next(providers.value.filter(e => e !== entry)) }
        },
        get: (id, params) =>
            providers.pipe(
                map(providers => providers.find(e => e.id === id)),
                distinctUntilChanged(),
                switchMap(e => (e ? e.provider(params) : of(null)))
            ),
    }
}
