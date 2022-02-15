import { MutableRefObject, useRef, useState } from 'react'
import { SpringValue, useSpring } from 'react-spring'

type UseAccordionResult<Element extends HTMLElement = HTMLDivElement> = [
    reference: MutableRefObject<Element | null>,
    open: boolean,
    toggleOpen: (open: boolean) => void,
    style: {
        height: SpringValue<string>
        opacity: SpringValue<number>
    }
]

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
