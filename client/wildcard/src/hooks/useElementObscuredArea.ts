import React from 'react'

import { throttle } from 'lodash'

interface ElementObscuredArea {
    top: number
    right: number
    bottom: number
    left: number
}

const SCROLL_THROTTLE_WAIT = 50

/**
 * Returns area obscured by scrolling of an element.
 */
export function useElementObscuredArea<T extends HTMLElement>(
    elementReference: React.MutableRefObject<T | null>,
    lazy?: boolean
): ElementObscuredArea {
    const [obscured, setObscured] = React.useState<ElementObscuredArea>({
        top: 0,
        right: 0,
        bottom: 0,
        left: 0,
    })

    const calculate = React.useMemo(
        () =>
            throttle(
                () => {
                    const element = elementReference?.current
                    if (element) {
                        setObscured({
                            top: Math.floor(element.scrollTop),
                            right: Math.floor(element.scrollWidth - element.clientWidth - element.scrollLeft),
                            bottom: Math.floor(element.scrollHeight - element.clientHeight - element.scrollTop),
                            left: Math.floor(element.scrollLeft),
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
            if (lazy) {
                scheduleIntoNextFrame(calculate)
            } else {
                calculate()
            }
            element.addEventListener('scroll', calculate, { passive: true })
        }
        return () => {
            element?.removeEventListener('scroll', calculate)
        }
    }, [elementReference, calculate, lazy])

    return obscured
}

function scheduleIntoNextFrame(callback: () => void): void {
    requestAnimationFrame(() => {
        requestAnimationFrame(() => {
            callback()
        })
    })
}
