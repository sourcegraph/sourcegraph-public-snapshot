import { useRef, useState } from 'react'

/**
 * A custom hook that tells if overflow: ellipsis is activated for a HTML element
 * For example, check if the text is currently truncated on mouse enter
 * We want to do it on mouse enter as browser window size might change after the element
 * has been loaded initially
 */
export function useIsTruncated<Element extends HTMLElement>(): {
    /**
     * Reference to a HTML element
     */
    elementReference: React.RefObject<Element>
    /**
     * Check if overflow: ellipsis is activated for the element
     */
    isTruncated: boolean
    /**
     * A function to check if overflow: ellipsis is currently
     * activated for the refered element, usually on mouse hover
     */
    onMouseEnter: () => void
} {
    const [isTruncated, setIsTruncated] = useState<boolean>(false)
    const elementReference = useRef<Element>(null)
    /**
     * check if ellipsis has been activated by comparing
     * the current client width and scroll width of the element
     * As the scroll width tells us the original width of the element
     * while the client width tells us the current viewpoint's width
     * of the element
     *  */
    function onMouseEnter(): void {
        if (elementReference.current) {
            setIsTruncated(elementReference.current.clientWidth < elementReference.current.scrollWidth)
        }
    }

    return {
        isTruncated,
        elementReference,
        onMouseEnter,
    }
}
