import { Placement, VirtualElement, Strategy, flip } from '@floating-ui/core'
import { getScrollParents, computePosition, shift, limitShift, offset } from '@floating-ui/dom'
import React, { forwardRef, useEffect, useLayoutEffect, useRef } from 'react'
import { createPortal } from 'react-dom'

export type Target = Element | VirtualElement

interface FloatingPanelProps extends React.HTMLAttributes<HTMLDivElement> {
    target: Target
    placement?: Placement
    strategy?: Strategy
    padding?: number
}

export function isElement(value: unknown): value is Element {
    return value instanceof window.Element
}

/**
 * Render floating panel element (tooltip, popover) according to target position,
 * parents scroll box rectangles, floating settings (like placement and target sizes)
 */
export const FloatingPanel: React.FunctionComponent<FloatingPanelProps> = props => {
    const { target, placement = 'right', strategy = 'absolute', children, padding = 10, ...otherProps } = props

    const floating = useRef<HTMLDivElement>(null)

    useEffect(() => {
        const floatingElement = floating.current

        if (!floatingElement) {
            return
        }

        async function update(): Promise<void> {
            if (!floatingElement) {
                return
            }

            const parents = [
                ...(isElement(target) ? getScrollParents(target) : []),
                ...getScrollParents(floatingElement),
            ] as Element[]

            const { x: xCoordinate, y: yCoordinate, middlewareData } = await computePosition(target, floatingElement, {
                placement,
                strategy,
                middleware: [
                    shift({ limiter: limitShift(), boundary: parents }),
                    offset(padding),
                    flip({ boundary: parents }),
                ],
            })

            const { referenceHidden } = middlewareData.hide ?? {}

            Object.assign(floatingElement.style, {
                position: strategy,
                top: 0,
                left: 0,
                visibility: referenceHidden ? 'hidden' : 'visible',
                transform: `translate(${Math.round(xCoordinate ?? 0)}px,${Math.round(yCoordinate ?? 0)}px)`,
            })
        }

        // Initial calculation on component mount
        // eslint-disable-next-line @typescript-eslint/no-floating-promises
        update()

        const parents = [...(isElement(target) ? getScrollParents(target) : []), ...getScrollParents(floatingElement)]

        for (const parent of parents) {
            parent.addEventListener('scroll', update)
            parent.addEventListener('resize', update)
        }

        return () => {
            for (const parent of parents) {
                parent.removeEventListener('scroll', update)
                parent.removeEventListener('resize', update)
            }
        }
    }, [floating, placement, strategy, target, padding])

    return (
        <FloatingPanelContent {...otherProps} portal={strategy === 'fixed'} ref={floating}>
            {children}
        </FloatingPanelContent>
    )
}

interface FloatingPanelContentProps extends React.HTMLAttributes<HTMLDivElement> {
    portal: boolean
}

const FloatingPanelContent = forwardRef<HTMLDivElement, FloatingPanelContentProps>((props, reference) => {
    const { portal, children, ...otherProps } = props

    const containerReference = useRef(document.createElement('div'))

    // Add a container element right after the body tag
    useLayoutEffect(() => {
        const element = containerReference.current

        if (!portal) {
            return
        }

        document.body.append(element)

        return () => {
            element.remove()
        }
    }, [containerReference, portal])

    return portal ? (
        createPortal(
            <div ref={reference} {...otherProps}>
                {children}
            </div>,
            containerReference.current
        )
    ) : (
        <div ref={reference} {...otherProps}>
            {children}
        </div>
    )
})
