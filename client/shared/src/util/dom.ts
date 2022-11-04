/**
 * Checks whether the element can receive (keyboard) input. That's the case for
 * <input>, <textarea> and content editable elements.
 */
export function isInputElement(element: Element): boolean {
    switch (element.nodeName) {
        case 'INPUT':
        case 'TEXTAREA':
            return true
        default:
            return elementIsContentEditable(element as HTMLElement)
    }
}

/**
 * Empty string and 'true' indicate that the element is editable. 'false'
 * indicates that the element is not editable. A missing value or an invalid
 * value implies inheritence from the parent.
 *
 * See https://html.spec.whatwg.org/multipage/interaction.html#attr-contenteditable
 */
function elementIsContentEditable(element: HTMLElement): boolean {
    switch (element.contentEditable) {
        case '':
        case 'true':
            return true
        case 'false':
            return false
        default:
            return element.parentElement ? elementIsContentEditable(element.parentElement) : false
    }
}

/**
 * Creates an SVG node. To be used together with path specs from @mdi/js
 */
export function createSVGIcon(pathSpec: string): Node {
    const svgNS = 'http://www.w3.org/2000/svg'
    const svg = document.createElementNS(svgNS, 'svg')
    svg.setAttributeNS(null, 'viewBox', '0 0 24 24')
    svg.setAttribute('aria-hidden', 'true')

    const path = document.createElementNS(svgNS, 'path')
    path.setAttribute('d', pathSpec)

    svg.append(path)
    return svg
}
