import Popper from 'popper.js'

import { Tooltip } from './constants'

import styles from './Tooltip.module.scss'

/**
 * Find the nearest ancestor element to e that contains a tooltip.
 */
export const getSubject = (element: HTMLElement | null): HTMLElement | null => {
    // If element is not actually attached to the DOM, then abort.
    if (!element || !document.body.contains(element)) {
        return null
    }

    return element.closest<HTMLElement>(`[${Tooltip.SUBJECT_ATTRIBUTE}]`)
}

export const getContent = (subject: HTMLElement): string | undefined =>
    subject.getAttribute(Tooltip.SUBJECT_ATTRIBUTE) || undefined

export const getPlacement = (subject: HTMLElement): Popper.Placement | undefined =>
    (subject.getAttribute(Tooltip.PLACEMENT_ATTRIBUTE) as Popper.Placement) || undefined

export const getDelay = (subject: HTMLElement): number | undefined => {
    const dataDelay = subject.getAttribute(Tooltip.DELAY_ATTRIBUTE)

    return dataDelay ? parseInt(dataDelay, 10) : undefined
}

/**
 * Sets or removes a plain-text tooltip on the HTML element using the native style for Sourcegraph
 * web app.
 *
 * @param element The HTML element whose tooltip to set or remove.
 * @param tooltip The tooltip plain-text content (to add the tooltip) or `null` (to remove the
 * tooltip).
 */
export function setElementTooltip(element: HTMLElement, tooltip: string | null): void {
    if (tooltip) {
        element.dataset.tooltip = tooltip
    } else {
        element.removeAttribute(Tooltip.SUBJECT_ATTRIBUTE)
    }
}

export const getTooltipStyle = (placement: Popper.Placement): string => {
    // values with start/end don't actually have a class.
    // if you check the css module.

    switch (placement) {
        case 'left':
        case 'left-end':
        case 'left-start':
            return styles.tooltipLeft
        case 'right':
        case 'right-end':
        case 'right-start':
            return styles.tooltipRight
        case 'bottom':
        case 'bottom-end':
        case 'bottom-start':
            return styles.tooltipBottom
        case 'top':
        case 'top-end':
        case 'top-start':
            return styles.tooltipTop
        default:
            return styles.tooltipAuto
    }
}
