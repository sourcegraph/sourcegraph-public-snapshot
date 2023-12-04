import React from 'react'

import { debounce } from 'lodash'

import { useReducedMotion } from '../../hooks'

import { useTabsState } from './context'

const SCROLL_BACK_WAIT = 500

export function useScrollBackToActive<T extends HTMLElement>(
    containerReference: React.MutableRefObject<T | null>
): void {
    const { activeIndex } = useTabsState()
    const isReducedMotion = useReducedMotion()

    const scrollBack = React.useMemo(
        () =>
            debounce(() => {
                if (containerReference?.current) {
                    containerReference.current.children.item(activeIndex)?.scrollIntoView({
                        behavior: isReducedMotion ? 'auto' : 'smooth',
                        inline: 'center',
                    })
                }
            }, SCROLL_BACK_WAIT),
        [activeIndex, containerReference, isReducedMotion]
    )

    React.useEffect(() => {
        const container = containerReference?.current
        const cancel: () => void = () => scrollBack.cancel()

        container?.addEventListener('mouseenter', cancel)
        container?.addEventListener('mouseleave', scrollBack)
        return () => {
            container?.removeEventListener('mouseenter', cancel)
            container?.removeEventListener('mouseleave', scrollBack)
        }
    }, [containerReference, scrollBack])
}
