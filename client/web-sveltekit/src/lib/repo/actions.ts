import { onDestroy, type ComponentType } from 'svelte'
import { writable, derived, type Readable } from 'svelte/store'

interface Action {
    key: string
    priority: number
    component: ComponentType
}

export interface ActionStore extends Readable<Action[]> {
    setAction(action: Action): void
}

/**
 * Creates a context via which repo subpages can add "actions" to the shared
 * header.
 */
export function createActionStore(): ActionStore {
    const actions = writable<Action[]>([])
    const sortedActions = derived(actions, $actions => [...$actions].sort((a, b) => a.priority - b.priority))

    // TODO: This should be reimplemented so that it's possible to add and
    // remove actions even after the component was instantiated.

    return {
        subscribe: sortedActions.subscribe,
        setAction(action: Action): void {
            actions.update(actions => {
                const existingAction = actions.find(a => a.key === action.key)
                if (existingAction) {
                    if (existingAction.component === action.component) {
                        return actions
                    }
                    actions = actions.filter(a => a.key !== action.key)
                }
                return [...actions, action]
            })
            onDestroy(() => {
                actions.update(actions => actions.filter(a => a.key !== action.key))
            })
        },
    }
}
