import { type MutableRefObject, useRef, useState } from 'react'

type UseIsTruncated<Element extends HTMLElement = HTMLDivElement> = [
    /**
     * Reference to a HTML element
     */
    reference: MutableRefObject<Element | null>,
    /**
     * Check if overflow: ellipsis is activated for the element
     */
    truncated: boolean,
    /**
     * A function to check if overflow: ellipsis is currently
     * activated for the refered element
     */
    checkTruncation: () => void
]

/**
 * A custom hook that tells if overflow: ellipsis is activated for a HTML element
 * For example, check if the text inside a div is currently truncated on mouse over
 */
export function useIsTruncated<Element extends HTMLElement = HTMLDivElement>(): UseIsTruncated<Element> {
    const elementReference = useRef<Element | null>(null)
    const [isTruncated, setIsTruncated] = useState<boolean>(false)
    /**
     * Check if ellipsis has been activated by comparing
     * the current client width and scroll width of the element
     * As the scroll width tells us the original width of the element
     * while the client width tells us the current viewpoint's width
     * of the element
     **/
    function checkIsTruncated(): void {
        if (elementReference.current) {
            setIsTruncated(elementReference.current.clientWidth < elementReference.current.scrollWidth)
        }
    }

    return [elementReference, isTruncated, checkIsTruncated]
}
