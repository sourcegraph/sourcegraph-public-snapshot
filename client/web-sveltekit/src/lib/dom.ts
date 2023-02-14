import type { ActionReturn } from 'svelte/action'

/**
 * An action that dispatches a custom 'click-outside' event when the user clicks
 * outside the attached element.
 */
export function onClickOutside(
    node: HTMLElement
): ActionReturn<void, { 'on:click-outside': (event: CustomEvent<HTMLElement>) => void }> {
    function handler(event: MouseEvent): void {
        if (event.target && !node.contains(event.target as HTMLElement)) {
            node.dispatchEvent(new CustomEvent('click-outside', { detail: event.target }))
        }
    }

    window.addEventListener('mousedown', handler)

    return {
        destroy() {
            window.removeEventListener('mousedown', handler)
        },
    }
}
