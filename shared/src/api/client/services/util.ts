import { compact, flatten } from 'lodash'

/** Flattens and compacts the argument. If it is null or if the result is empty, it returns null. */
export function flattenAndCompact<T>(value: (T | T[] | undefined | null)[] | null): T[] | null {
    if (value === null || value === undefined) {
        return null
    }
    const merged = flatten(compact(value))
    return merged.length === 0 ? null : merged
}
