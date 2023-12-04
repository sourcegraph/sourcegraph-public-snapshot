import { type MutableRefObject, useRef, useState } from 'react'

import { type SpringValue, useSpring } from 'react-spring'

type UseAccordionResult<Element extends HTMLElement = HTMLDivElement> = [
    reference: MutableRefObject<Element | null>,
    open: boolean,
    toggleOpen: (open: boolean) => void,
    style: {
        height: SpringValue<string>
        opacity: SpringValue<number>
    }
]

/**
 * Custom hook which can animate a collapsible "accordion" element with an automatic
 * height. Returns a tuple with:
 * - A React `MutableRefObject` which should be attached to the panel contents.
 * - A boolean value which indicates whether the accordion is open or not.
 * - A function which toggles the accordion open/closed.
 * - A style object which applies an animated height and opacity to the contents.
 *
 * @param defaultOpen whether or not the accordion element should start open/expanded
 */
export const useAccordion = <Element extends HTMLElement = HTMLDivElement>(
    defaultOpen = false
): UseAccordionResult<Element> => {
    const panelReference = useRef<Element | null>(null)
    const [exampleOpen, setExampleOpen] = useState(defaultOpen)
    const exampleStyle = useSpring({
        height: exampleOpen ? `${panelReference.current?.offsetHeight || '100'}px` : '0px',
        opacity: exampleOpen ? 1 : 0,
    })

    return [panelReference, exampleOpen, setExampleOpen, exampleStyle]
}
