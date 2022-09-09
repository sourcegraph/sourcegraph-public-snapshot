import { useEffect } from 'react'

import { debounce } from 'lodash'
import { useHistory, useLocation } from 'react-router'

/**
 * Object containing maps of saved scroll locations for each path name, which are keyed by the scrollable container they belong to.
 * This will allow us to easily support scroll preservation in multiple containers, even when they are used in the same page.
 */
const SAVED_SCROLL_POSITIONS: { [k: string]: Map<string, number> } = {}

/**
 * Object containing maps of cached scroll attempts for each path name, which are keyed by the scrollable container they belong to.
 * This can be used to cancel a scroll attempt if a new attempt comes in for the same path.
 */
const CACHED_SCROLL_ATTEMPTS: { [k: string]: Map<string, MutationObserverPromise> } = {}

/** The length of the timeout in milliseconds until we stop trying to scroll. */
const SCROLL_RETRY_TIMEOUT = 3000

export function useScrollManager(containerKey: string, containerRef: React.RefObject<HTMLElement>): void {
    const { pathname } = useLocation()
    const { action } = useHistory()

    // Set up the maps for this containerKey if they haven't been created yet
    useEffect(() => {
        if (!SAVED_SCROLL_POSITIONS[containerKey]) {
            SAVED_SCROLL_POSITIONS[containerKey] = new Map<string, number>()
        }

        if (!CACHED_SCROLL_ATTEMPTS[containerKey]) {
            CACHED_SCROLL_ATTEMPTS[containerKey] = new Map<string, MutationObserverPromise>()
        }
    }, [containerKey])

    // Attach scroll listener, make sure to detach on unmount
    useEffect(() => {
        const container = containerRef.current

        const saveScrollPosition = debounce(() => {
            if (container) {
                console.log(`Saving position '${container.scrollTop}' for pathname '${pathname}'`)
                SAVED_SCROLL_POSITIONS[containerKey].set(pathname, container.scrollTop)
            }
        }, 200)

        container?.addEventListener('scroll', saveScrollPosition)

        return () => container?.removeEventListener('scroll', saveScrollPosition)
    }, [pathname, containerKey, containerRef])

    // Handle changes to pathname and try to scroll to the saved position at that path, if it exists
    useEffect(() => {
        // Cancel any existing cached scroll attempts for this pathname
        if (CACHED_SCROLL_ATTEMPTS[containerKey].has(pathname)) {
            CACHED_SCROLL_ATTEMPTS[containerKey].get(pathname)?.cancel()
            CACHED_SCROLL_ATTEMPTS[containerKey].delete(pathname)
        }

        // Attempt a scroll if we have a saved position for the pathname; if the scroll doesn't work, set up an observer to
        // retry up until a given timeout
        if (action === 'POP' && SAVED_SCROLL_POSITIONS[containerKey].has(pathname)) {
            const scrollPosition = SAVED_SCROLL_POSITIONS[containerKey].get(pathname) ?? 0

            const attemptScroll = (): boolean => {
                containerRef.current?.scrollTo(0, scrollPosition)
                return containerRef.current?.scrollTop === scrollPosition
            }

            if (!attemptScroll()) {
                const promiseToCache = observerPromiseWithTimeout(
                    attemptScroll,
                    SCROLL_RETRY_TIMEOUT,
                    containerRef.current
                )
                CACHED_SCROLL_ATTEMPTS[containerKey].set(pathname, promiseToCache)

                promiseToCache
                    .then(() =>
                        // Delete the cached scroll for this path once it resolves
                        CACHED_SCROLL_ATTEMPTS[containerKey].delete(pathname)
                    )
                    .catch((error: MutationObserverError) => {
                        if (!error.cancelled) {
                            console.error(
                                `Failed to scroll to position '${scrollPosition}' for pathname '${pathname}' in container '${containerKey}'.`
                            )
                        }
                    })
            }
        }
        // This genuinely should only run when pathname changes, other dependencies should not change
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [pathname])
}

class MutationObserverPromise extends Promise<unknown> {
    public cancel: () => void = () => {}
}

class MutationObserverError extends Error {
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
function observerPromiseWithTimeout(
    callback: () => boolean,
    timeout: number,
    node: HTMLElement | null
): MutationObserverPromise {
    let cancel: () => void

    const result = new MutationObserverPromise((resolve, reject) => {
        let success: boolean

        const observer = buildMutationObserver(() => {
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
