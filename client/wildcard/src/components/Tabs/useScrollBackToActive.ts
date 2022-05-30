import React from 'react'

import { useMatchMedia } from '@sourcegraph/wildcard'

import { useTabsState } from './context'

import { debounce } from 'lodash'

const SCROLL_BACK_WAIT = 500

export function useScrollBackToActive<T extends HTMLElement>(
    containerReference: React.MutableRefObject<T | null>
): void {
    const { activeIndex } = useTabsState()
    const isReducedMotion = useMatchMedia('(prefers-reduced-motion: reduce)')

    const scrollBack = React.useMemo(() => {
        return debounce(() => {
            if (containerReference?.current) {
                containerReference.current.children.item(activeIndex)?.scrollIntoView({
                    behavior: isReducedMotion ? 'auto' : 'smooth',
                    inline: 'center',
                })
            }
        }, SCROLL_BACK_WAIT)
    }, [activeIndex, containerReference, isReducedMotion])

    const cancelScrollBack = React.useCallback(() => {
        scrollBack.cancel()
    }, [scrollBack])

    React.useEffect(() => {
        const container = containerReference?.current
        container?.addEventListener('mouseenter', cancelScrollBack)
        container?.addEventListener('mouseleave', scrollBack)
        return () => {
            container?.removeEventListener('mouseenter', cancelScrollBack)
            container?.removeEventListener('mouseleave', scrollBack)
        }
    }, [containerReference, scrollBack, cancelScrollBack])
}
