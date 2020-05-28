/**
 * limitString limits the given string to N characters, optionally adding an
 * ellipsis (…) at the end.
 *
 * @param s the string to limit
 * @param n the number of characters to limit the string to
 * @param ellipsis whether or not to add an ellipsis (…) when string is cut off.
 */
export function limitString(s: string, n: number, ellipsis: boolean): string {
    if (s.length > n) {
        if (ellipsis) {
            return s.slice(0, n - 1) + '…'
        }
        return s.slice(0, n)
    }
    return s
}

/**
 * scrollIntoView checks if the selected element is not in view of the list
 * element, adjusting the scroll of the list element as needed.
 *
 * @param listElement the list element.
 * @param selectedElement the selected element.
 */
export function scrollIntoView(listElement?: HTMLElement, selectedElement?: HTMLElement): void {
    if (!listElement || !selectedElement) {
        return
    }
    const listRect = listElement.getBoundingClientRect()
    const selectedRect = selectedElement.getBoundingClientRect()

    if (selectedRect.top <= listRect.top) {
        // Selected item is out of view at the top of the list.
        listElement.scrollTop -= listRect.top - selectedRect.top
    } else if (selectedRect.bottom >= listRect.bottom) {
        // Selected item is out of view at the bottom of the list.
        listElement.scrollTop += selectedRect.bottom - listRect.bottom
    }
}
