export type NextSignature<P, R> = (this: void, data: P, next: (data: P) => R) => R
