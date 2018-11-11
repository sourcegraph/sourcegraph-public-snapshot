/**
 * This file contains (simplified) implementations of Lodash functions. We use these instead of depending on Lodash
 * because depending on it (1) either results in a significantly larger bundle size, if tree-shaking is not enabled
 * (and the npm lodash package is used), or (2) significantly increases the complexity of bundling and executing
 * code, if tree-shaking is enabled (and the npm lodash-es package is used).
 */
/** Flattens the array one level deep. */
export declare function flatten<T>(array: (T | T[])[]): T[];
/** Removes all falsey values. */
export declare function compact<T>(array: (T | null | undefined | false | '' | 0)[]): T[];
/** Reports whether the two values are equal, using a strict deep comparison. */
export declare function isEqual<T>(a: T, b: T): boolean;
/**
 * Runs f and returns a resolved promise with its value or a rejected promise with its exception,
 * regardless of whether it returns a promise or not.
 */
export declare function tryCatchPromise<T>(f: () => T | Promise<T>): Promise<T>;
