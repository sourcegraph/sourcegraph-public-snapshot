import { Position } from 'vscode-languageserver-types'
import { HoverMerged } from '../../backend/features'
import { isEmptyHover } from '../../backend/lsp'
import { LineOrPositionOrRange } from '../../util/url'
import { findElementWithOffset } from '../blob/tooltips'
import { HoverOverlayProps, isJumpURL } from './HoverOverlay'

export const LOADING: 'loading' = 'loading'

/**
 * @param codeElement The `<code>` element
 * @param line 1-indexed line number
 * @return The `<tr>` element
 */
export const getRowInCodeElement = (codeElement: HTMLElement, line: number): HTMLTableRowElement | undefined => {
    const table = codeElement.firstElementChild as HTMLTableElement
    return table.rows[line - 1]
}

/**
 * Returns a list of `<tr>` elements that are contained in the given range
 *
 * @param position 1-indexed line, position or inclusive range
 */
export const getRowsInRange = (
    codeElement: HTMLElement,
    position?: LineOrPositionOrRange
): {
    /** 1-indexed line number */
    line: number
    /** The `<tr>` element */
    element: HTMLTableRowElement
}[] => {
    if (!position || position.line === undefined) {
        return []
    }
    const tableElement = codeElement.firstElementChild as HTMLTableElement
    const rows: { line: number; element: HTMLTableRowElement }[] = []
    for (let line = position.line; line <= (position.endLine || position.line); line++) {
        const element = tableElement.rows[line - 1]
        if (!element) {
            break
        }
        rows.push({ line, element })
    }
    return rows
}

/**
 * Returns the token `<span>` element in a `<code>` element for a given 1-indexed position.
 *
 * @param codeElement The `<code>` element
 * @param position 1-indexed position
 */
export const getTokenAtPosition = (codeElement: HTMLElement, position: Position): HTMLElement | undefined => {
    const row = getRowInCodeElement(codeElement, position.line)
    if (!row) {
        return undefined
    }
    const [, codeCell] = row.cells
    return findElementWithOffset(codeCell, position.character)
}

/**
 * `padding-top` of the blob element in px.
 * TODO find a way to remove the need for this.
 */
export const BLOB_PADDING_TOP = 8

/**
 * Calculates the desired position of the hover overlay depending on the container,
 * the hover target and the size of the hover overlay
 *
 * @param scrollable The closest container that is scrollable
 * @param target The DOM Node that was hovered
 * @param tooltip The DOM Node of the tooltip
 */
export const calculateOverlayPosition = (
    scrollable: HTMLElement,
    target: HTMLElement,
    tooltip: HTMLElement
): { left: number; top: number } => {
    // The scrollable element is the one with scrollbars. The scrolling element is the one with the content.
    const scrollableBounds = scrollable.getBoundingClientRect()
    const targetBound = target.getBoundingClientRect() // our target elements bounds

    // Anchor it horizontally, prior to rendering to account for wrapping
    // changes to vertical height if the tooltip is at the edge of the viewport.
    const relLeft = targetBound.left - scrollableBounds.left

    // Anchor the tooltip vertically.
    const tooltipBound = tooltip.getBoundingClientRect()
    const relTop = targetBound.top + scrollable.scrollTop - scrollableBounds.top
    // This is the padding-top of the blob element
    let tooltipTop = relTop - (tooltipBound.height - BLOB_PADDING_TOP)
    if (tooltipTop - scrollable.scrollTop < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = targetBound.bottom + scrollable.scrollTop - scrollableBounds.top
        tooltipTop = relBottom
    } else {
        tooltipTop -= BLOB_PADDING_TOP
    }
    return { left: relLeft, top: tooltipTop }
}

/**
 * Scrolls an element to the center if it is out of view.
 * Does nothing if the element is in view.
 *
 * @param container The scrollable container (that has `overflow: auto`)
 * @param content The content child that is being scrolled
 * @param target The element that should be scrolled into view
 */
export const scrollIntoCenterIfNeeded = (container: HTMLElement, content: HTMLElement, target: HTMLElement): void => {
    const blobRect = container.getBoundingClientRect()
    const rowRect = target.getBoundingClientRect()
    if (rowRect.top <= blobRect.top || rowRect.bottom >= blobRect.bottom) {
        const blobRect = container.getBoundingClientRect()
        const contentRect = content.getBoundingClientRect()
        const rowRect = target.getBoundingClientRect()
        const scrollTop = rowRect.top - contentRect.top - blobRect.height / 2 + rowRect.height / 2
        container.scrollTop = scrollTop
    }
}

/**
 * Returns true if the HoverOverlay would have anything to show according to the given hover and definition states.
 */
export const overlayUIHasContent = (state: Pick<HoverOverlayProps, 'hoverOrError' | 'definitionURLOrError'>): boolean =>
    (state.hoverOrError && !(HoverMerged.is(state.hoverOrError) && isEmptyHover(state.hoverOrError))) ||
    isJumpURL(state.definitionURLOrError)
