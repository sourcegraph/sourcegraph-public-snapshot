import React from 'react'

import { throttle } from 'lodash'

interface ElementObscuredArea {
    left: number
    right: number
}

const SCROLL_THROTTLE_WAIT = 50

/**
 * Returns area obscured by scrolling of an element.
 */
export function useElementObscuredArea<T extends HTMLElement>(
    elementReference: React.MutableRefObject<T | null>
): ElementObscuredArea {
    const [obscured, setObscured] = React.useState<ElementObscuredArea>({
        left: 0,
        right: 0,
    })

    const calculate = React.useMemo(
        () =>
            throttle(
                () => {
                    const element = elementReference?.current
                    if (element) {
                        setObscured({
                            left: element.scrollLeft,
                            right: element.scrollWidth - element.clientWidth - element.scrollLeft,
                        })
                    }
                },
                SCROLL_THROTTLE_WAIT,
                { leading: true, trailing: true }
            ),
        [elementReference]
    )

    React.useEffect(() => {
        const element = elementReference?.current
        if (element) {
            calculate()
            element.addEventListener('scroll', calculate, { passive: true })
        }
        return () => {
            element?.removeEventListener('scroll', calculate)
        }
    }, [elementReference, calculate])

    return obscured
}
