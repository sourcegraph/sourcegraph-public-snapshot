import { once } from 'lodash'
import type { ActionReturn } from 'svelte/action'

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

const getGlobalObserver = once(() => createObserver({ root: null, rootMargin: '0px 0px 500px 0px' }))

export function observeIntersection(
    node: HTMLElement
): ActionReturn<void, { 'on:intersecting': (e: CustomEvent<boolean>) => void }> {
    // If the environment doesn't support IntersectionObserver we assume that the
    // element is visible and dispatch the event immediately
    if (!supportsIntersectionObserver()) {
        node.dispatchEvent(new CustomEvent<boolean>('intersecting', { detail: true }))
        return {}
    }

    let observer = getGlobalObserver()

    let scrollAncestor: HTMLElement | null = node.parentElement
    while (scrollAncestor) {
        const overflow = getComputedStyle(scrollAncestor).overflowY
        if (overflow === 'auto' || overflow === 'scroll') {
            break
        }
        scrollAncestor = scrollAncestor.parentElement
    }

    if (scrollAncestor && scrollAncestor !== document.getRootNode()) {
        observer = createObserver({ root: scrollAncestor, rootMargin: '0px 0px 500px 0px' })
    }

    observer.observe(node)

    return {
        destroy() {
            observer.unobserve(node)
        },
    }
}
