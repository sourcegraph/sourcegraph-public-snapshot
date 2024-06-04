/**
 * Checks whether the element can receive (keyboard) input. That's the case for
 * <input>, <textarea> and content editable elements.
 */
export function isInputElement(element: Element): boolean {
    switch (element.nodeName) {
        case 'INPUT':
        case 'TEXTAREA': {
            return true
        }
        default: {
            return elementIsContentEditable(element as HTMLElement)
        }
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
        case 'true': {
            return true
        }
        case 'false': {
            return false
        }
        default: {
            return element.parentElement ? elementIsContentEditable(element.parentElement) : false
        }
    }
}

const SVG_NAMESPACE = 'http://www.w3.org/2000/svg'
const HTML_NAMESPACE = 'http://www.w3.org/1999/xhtml'

/**
 * Creates an SVG node. To be used together with path specs from @mdi/js
 */
export function createSVGIcon(pathSpec: string, ariaLabel?: string): SVGElement {
    const svg = document.createElementNS(SVG_NAMESPACE, 'svg')
    svg.style.fill = 'currentcolor'
    svg.setAttribute('viewBox', '0 0 24 24')
    if (ariaLabel) {
        svg.setAttributeNS(HTML_NAMESPACE, 'aria-label', ariaLabel)
    } else {
        svg.setAttributeNS(HTML_NAMESPACE, 'aria-hidden', 'true')
    }

    const path = document.createElementNS(SVG_NAMESPACE, 'path')
    path.setAttribute('d', pathSpec)
    svg.append(path)

    return svg
}

/**
 * Helper function for creating DOM elements with properties and children.
 *
 * @param tagName The tag name of the element to create.
 * @param properties The properties to set on the element.
 * @param children The children to append to the element.
 * @returns The created element.
 */
export function createElement<K extends keyof HTMLElementTagNameMap>(
    tagName: K,
    properties: Partial<HTMLElementTagNameMap[K]> | null = null,
    ...children: (Node | string)[]
): HTMLElementTagNameMap[K] {
    const element = Object.assign(document.createElement(tagName), properties)
    for (const child of children) {
        element.append(typeof child === 'string' ? document.createTextNode(child) : child)
    }
    return element
}
