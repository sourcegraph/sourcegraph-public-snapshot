import { createPopper, type Instance, type Options } from '@popperjs/core'
import type { ActionReturn, Action } from 'svelte/action'
import * as uuid from 'uuid'

import { highlightNode } from '$lib/common'

/**
 * Returns a unique ID to be used with accessible elements.
 * Generates stable IDs in tests.
 */
export function uniqueID(prefix = '') {
    if (process.env.VITEST) {
        return `test-${prefix}-123`
    }
    return `${prefix}-${uuid.v4()}`
}

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

interface PopperReturnValue {
    /**
     * Force update positioning and layout of the popover.
     */
    update: () => void

    /**
     * Attach this action to the element that represents the popover content.
     * "Target" is the element that triggers the popover.
     */
    popover: Action<HTMLElement, { target: Element; options: Partial<Options> }>
}

/**
 * Returns an action that converts an element into a popover.
 */
export function createPopover(): PopperReturnValue {
    let popperInstance: Instance | null
    return {
        update: () => popperInstance?.update(),
        popover: (node, { target, options }) => {
            popperInstance = createPopper(target, node, options)

            return {
                update(parameter) {
                    if (parameter.target !== target) {
                        popperInstance?.destroy()
                        popperInstance = createPopper(parameter.target, node, parameter.options)
                    } else {
                        popperInstance?.setOptions(parameter.options)
                        popperInstance?.update()
                    }
                },
                destroy() {
                    popperInstance?.destroy()
                    popperInstance = null
                },
            }
        },
    }
}

/**
 * Updates the DOM to highlight the provided ranges.
 * IMPORTANT: If the element content is dynamic you have to ensure that the attached is recreated
 * to properly update and re-highlight the content. One way to enforce this is to use #key
 */
export const highlightRanges: Action<HTMLElement, { ranges: [number, number][] }> = (node, parameters) => {
    function highlight({ ranges }: { ranges: [number, number][] }) {
        if (ranges.length > 0) {
            for (const [start, end] of ranges) {
                highlightNode(node, start, end - start)
            }
        }
    }

    highlight(parameters)

    return {
        update: highlight,
    }
}
