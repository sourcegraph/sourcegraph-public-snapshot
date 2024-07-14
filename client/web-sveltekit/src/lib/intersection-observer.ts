import type { Action } from 'svelte/action'

/**
 * Returns true if the environment supports IntersectionObserver
 * (usually not the case in a test environment.
 */
function supportsIntersectionObserver(): boolean {
    return !!globalThis.IntersectionObserver
}

function intersectionHandler(entries: IntersectionObserverEntry[]): void {
    for (const entry of entries) {
        entry.target.dispatchEvent(new CustomEvent<boolean>('intersecting', { detail: entry.isIntersecting }))
    }
}

function createObserver(init: IntersectionObserverInit): IntersectionObserver {
    return new IntersectionObserver(intersectionHandler, init)
}

const observerCache = new WeakMap<HTMLElement, IntersectionObserver>()
function getObserver(container: HTMLElement): IntersectionObserver {
    let observer = observerCache.get(container)
    if (!observer) {
        observer = createObserver({ root: container, rootMargin: '0px 0px 500px 0px' })
        observerCache.set(container, observer)
    }
    return observer
}

export const observeIntersection: Action<
    HTMLElement,
    HTMLElement | null,
    { 'on:intersecting': (e: CustomEvent<boolean>) => void }
> = (node: HTMLElement, container: HTMLElement | null) => {
    // If the environment doesn't support IntersectionObserver we assume that the
    // element is visible and dispatch the event immediately
    if (!supportsIntersectionObserver()) {
        node.dispatchEvent(new CustomEvent<boolean>('intersecting', { detail: true }))
        return {}
    }

    let observer = container ? getObserver(container) : null
    observer?.observe(node)

    return {
        update(newContainer) {
            container && observer?.unobserve(container)
            container = newContainer

            observer = container ? getObserver(container) : null
            observer?.observe(node)
        },
        destroy() {
            observer?.unobserve(node)
        },
    }
}
