import { WebAppEvent } from './events'

export function createDispatcher<Event, Store extends (event: Event) => void = (event: Event) => void>(): {
    dispatch: (event: Event) => void
    register: (store: Store) => void
} {
    const stores: Store[] = []

    return {
        dispatch(event: Event) {
            for (const store of stores) {
                store(event)
            }
        },
        register(store: Store) {
            stores.push(store)
        },
    }
}

/**
 * Default dispatcher for the web app
 */
export const { dispatch, register } = createDispatcher<WebAppEvent>()
