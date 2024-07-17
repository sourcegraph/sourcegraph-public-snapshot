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

/**
 * observeIntersection emits an `intersecting` event when the node intersects with the
 * target element. In the case that the target element is null, we fall back to intersection
 * with the root element.
 */
export const observeIntersection: Action<
    HTMLElement,
    HTMLElement | null,
    { 'on:intersecting': (e: CustomEvent<boolean>) => void }
> = (node: HTMLElement, target: HTMLElement | null) => {
    // If the environment doesn't support IntersectionObserver we assume that the
    // element is visible and dispatch the event immediately
    if (!supportsIntersectionObserver()) {
        node.dispatchEvent(new CustomEvent<boolean>('intersecting', { detail: true }))
        return {}
    }

    let container = target ?? document.documentElement
    let observer = getObserver(container)
    observer.observe(node)

    return {
        update(newContainer) {
            observer.unobserve(container)
            container = newContainer ?? document.documentElement

            observer = getObserver(container)
            observer.observe(node)
        },
        destroy() {
            observer.unobserve(node)
        },
    }
}
