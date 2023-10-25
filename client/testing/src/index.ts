export * from './aria-asserts'

/**
 * Returns a {@link Promise} and a function. The {@link Promise} blocks until the returned function is called.
 */
export function createBarrier(): { wait: Promise<void>; done: () => void } {
    let done!: () => void
    const wait = new Promise<void>(resolve => (done = resolve))
    return { wait, done }
}
