/**
 * The following code was adapted from react-scroll-manager: https://github.com/trevorr/react-scroll-manager/blob/master/src/timedMutationObserver.js
 *
 * react-scroll-manager is available under the ISC License:
 *
 * ISC License
 *
 * Copyright (c) 2018, Trevor Robinson
 *
 * Permission to use, copy, modify, and/or distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 */

export class MutationObserverPromise extends Promise<unknown> {
    public cancel: () => void = () => {}
}

export class MutationObserverError extends Error {
    public cancelled = false
    public timedOut = false
}

/**
 * Wraps a MutationObserver with a cancellable promise that will reject if the specified timeout is reached or
 * the promise is canceled.
 *
 * @param callback The function that will be used to mutate the DOM and check for success.
 * @param timeout How long to attempt retrying this mutation.
 * @param node The node that will be observed.
 */
export function mutationObserverWithTimeout(
    callback: () => boolean,
    timeout: number,
    node: HTMLElement | null
): MutationObserverPromise {
    let cancel: () => void = () => {}

    const result = new MutationObserverPromise((resolve, reject) => {
        let success: boolean

        const observer = buildMutationObserver(() => {
            // If we weren't already successful, try and run the callback again and see if it's successful now
            // eslint-disable-next-line callback-return
            if (!success && (success = callback())) {
                cancel()
                resolve(success)
            }
        }, node ?? document)

        cancel = () => {
            observer.disconnect()
            clearTimeout(timeoutId)
            if (!success) {
                const reason = new MutationObserverError('MutationObserver cancelled')
                reason.cancelled = true
                reject(reason)
            }
        }

        const timeoutId = setTimeout(() => {
            observer.disconnect()
            clearTimeout(timeoutId)
            if (!success) {
                const reason = new MutationObserverError('MutationObserver timed out')
                reason.timedOut = true
                reject(reason)
            }
        }, timeout)
    })

    result.cancel = cancel
    return result
}

/** Sets up a MutationObserver and beings observing. */
function buildMutationObserver(callback: () => void, node: HTMLElement | Document): MutationObserver {
    const observer = new MutationObserver(callback)
    observer.observe(node, {
        attributes: true,
        childList: true,
        subtree: true,
    })
    return observer
}
