import { getContext, setContext } from 'svelte'

export function createContextAccessors<T>(): [(context: T) => void, () => T] {
    const KEY = {}
    return [context => setContext<T>(KEY, context), () => getContext<T>(KEY)]
}
