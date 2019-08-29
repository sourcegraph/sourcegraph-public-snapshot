import React, { useCallback, useState } from 'react'

export interface StickyTopComponentProps {
    /**
     * Whether the component wrapped with {@link WithStickyTop} is stuck to the top of its
     * containing block. This occurs when the user scrolls the page down enough so that the
     * component's element would otherwise be out of the viewport.
     */
    isStuck: boolean
}

/**
 * A wrapper that renders its children with CSS `position: sticky` at the top of its containing
 * block *and* passes a prop for the component to know whether it is "stuck". See
 * https://developers.google.com/web/updates/2017/09/sticky-headers for the technique used.
 */
export const WithStickyTop: React.FunctionComponent<{
    scrollContainerSelector: string
    children: (props: StickyTopComponentProps) => React.ReactElement | null
}> = ({ scrollContainerSelector, children }) => {
    const [isStuck, setIsStuck] = useState(false)

    const setSentinelTop = useCallback((sentinelTop: HTMLElement | null) => {
        if (!sentinelTop) {
            return
        }

        // Find scrolling container.
        const container = sentinelTop.closest(scrollContainerSelector)
        if (!container) {
            setIsStuck(false)
            console.error('WithStickyTop: scrolling container not found')
            return
        }

        const observer = new IntersectionObserver(
            records => {
                for (const record of records) {
                    const targetInfo = record.boundingClientRect
                    const rootBoundsInfo = record.rootBounds

                    // Started sticking.
                    if (targetInfo.bottom < rootBoundsInfo.top) {
                        setIsStuck(true)
                    }

                    // Stopped sticking.
                    if (targetInfo.bottom >= rootBoundsInfo.top && targetInfo.bottom < rootBoundsInfo.bottom) {
                        setIsStuck(false)
                    }
                }
            },
            { threshold: [0], root: container }
        )

        observer.observe(sentinelTop)
    }, [])

    return (
        <>
            <div
                className="with-sticky-top__sentinel-top"
                ref={setSentinelTop}
                // tslint:disable-next-line: jsx-ban-props
                style={{
                    position: 'absolute',
                    left: 0,
                    right: 0,
                    visibility: 'hidden',
                }}
            />
            {children({ isStuck })}
        </>
    )
}
