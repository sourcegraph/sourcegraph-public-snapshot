import type { Action } from 'svelte/action'

export const scrollIntoView: Action<HTMLElement, boolean> = (node: HTMLElement, scroll: boolean) => {
    if (scroll) {
        setTimeout(() => node.scrollIntoView({ block: 'center' }), 0)
    }
}
