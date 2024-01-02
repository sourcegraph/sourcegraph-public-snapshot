import { createPopper, type Instance, type Options } from '@popperjs/core'
import type { ActionReturn, Action } from 'svelte/action'
import * as uuid from 'uuid'

import { highlightNode } from '$lib/common'

/**
 * Returns a unique ID to be used with accessible elements.
 * Generates stable IDs in tests.
 */
export function uniqueID(prefix = ''): string {
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

/**
 * This action ensures that the element does not extend outside the visible viewport,
 * by settings its max height. Only works for elements whose position doesn't change
 * relative to the viewport.
 */
export const restrictToViewport: Action<HTMLElement, { offset?: number }> = (node, parameters) => {
    let offset = parameters.offset ?? 0

    function setMaxHeight(): void {
        node.style.maxHeight = window.innerHeight - node.getBoundingClientRect().top + offset + 'px'
    }

    window.addEventListener('resize', setMaxHeight)

    setMaxHeight()

    return {
        update(parameter) {
            offset = parameter.offset ?? 0
            setMaxHeight()
        },

        destroy() {
            window.removeEventListener('resize', setMaxHeight)
        },
    }
}

/**
 * An action to compute the number of elements that fit inside the container.
 * This works by caching the position of the right hand of each child element,
 * and then comparing it to the right hand of the container.
 *
 * Because of this this action only works for static element lists,
 * i.e. on initial render the node needs to contain all possible child elements.
 */
export const computeFit: Action<HTMLElement> = (
    node
): ActionReturn<void, { 'on:fit': (event: CustomEvent<{ itemCount: number }>) => void }> => {
    // Holds the cumulative width of all elements up to element i.
    const widths: number[] = [0]

    for (let i = 0; i < node.children.length; i++) {
        widths[i + 1] = node.children[i].getBoundingClientRect().right
    }

    function compute(): void {
        const right = node.getBoundingClientRect().right
        for (let i = widths.length - 1; i >= 0; i--) {
            if (widths[i] < right) {
                node.dispatchEvent(new CustomEvent('fit', { detail: { itemCount: i } }))
                return
            }
        }
    }

    compute()
    const observer = new ResizeObserver(compute)
    observer.observe(node)

    return {
        destroy() {
            observer.disconnect()
        },
    }
}
