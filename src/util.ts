/**
 * This file contains (simplified) implementations of Lodash functions. We use these instead of depending on Lodash
 * because depending on it (1) either results in a significantly larger bundle size, if tree-shaking is not enabled
 * (and the npm lodash package is used), or (2) significantly increases the complexity of bundling and executing
 * code, if tree-shaking is enabled (and the npm lodash-es package is used).
 */

/** Flattens the array one level deep. */
export function flatten<T>(array: (T | T[])[]): T[] {
    const result: T[] = []
    for (const value of array) {
        if (Array.isArray(value)) {
            result.push(...value)
        } else {
            result.push(value)
        }
    }
    return result
}

/** Removes all falsey values. */
export function compact<T>(array: (T | null | undefined | false | '' | 0)[]): T[] {
    const result: T[] = []
    for (const value of array) {
        if (value) {
            result.push(value)
        }
    }
    return result
}

/** Reports whether the two values are equal, using a strict deep comparison. */
export function isEqual<T>(a: T, b: T): boolean {
    if (a === b) {
        return true
    }
    // tslint:disable-next-line:triple-equals
    if (!a || !b || (typeof a !== 'object' && typeof b !== 'object')) {
        return a === b
    }
    return equalObjects(a, b)
}

function equalObjects<T extends { [key: string]: any }>(a: T, b: T): boolean {
    const ka = Object.keys(a)
    const kb = Object.keys(b)
    if (ka.length !== kb.length) {
        return false
    }
    ka.sort()
    kb.sort()
    for (let i = ka.length - 1; i >= 0; i--) {
        if (ka[i] !== kb[i]) {
            return false
        }
    }
    for (let i = ka.length - 1; i >= 0; i--) {
        const key = ka[i]
        if (!isEqual(a[key], b[key])) {
            return false
        }
    }
    return typeof a === typeof b
}

/**
 * Runs f and returns a resolved promise with its value or a rejected promise with its exception,
 * regardless of whether it returns a promise or not.
 */
export function tryCatchPromise<T>(f: () => T | Promise<T>): Promise<T> {
    try {
        return Promise.resolve(f())
    } catch (err) {
        return Promise.reject(err)
    }
}
