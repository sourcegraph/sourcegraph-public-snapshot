import React from 'react'

import { useMatchMedia } from '@sourcegraph/wildcard'

import { useTabsState } from './context'

export function useScrollBackToActive<T extends HTMLElement>(
    containerReference: React.MutableRefObject<T | null>
): void {
    const { activeIndex } = useTabsState()
    const isReducedMotion = useMatchMedia('(prefers-reduced-motion: reduce)')

    const scrollBack = React.useCallback(() => {
        if (containerReference?.current) {
            containerReference.current.children.item(activeIndex)?.scrollIntoView({
                behavior: isReducedMotion ? 'auto' : 'smooth',
                inline: 'center',
            })
        }
    }, [activeIndex, containerReference, isReducedMotion])

    React.useEffect(() => {
        const container = containerReference?.current
        container?.addEventListener('mouseleave', scrollBack)
        return () => {
            container?.removeEventListener('mouseleave', scrollBack)
        }
    }, [containerReference, scrollBack])
}
