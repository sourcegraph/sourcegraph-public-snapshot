import { Placement, Strategy } from '@floating-ui/core'
import { Options as OffsetOptions } from '@floating-ui/core/src/middleware/offset'
import { computePosition, Middleware } from '@floating-ui/dom'
import React, { forwardRef, useLayoutEffect, useRef } from 'react'
import { createPortal } from 'react-dom'

import { getPositionMiddlewares, Target } from './utils'

interface FloatingPanelProps extends React.HTMLAttributes<HTMLDivElement> {
    target: Target
    placement?: Placement
    strategy?: Strategy
    padding?: OffsetOptions
    constraints?: Element[]
    middlewares?: Middleware[]
}

/**
 * Render floating panel element (tooltip, popover) according to target position,
 * parents scroll box rectangles, floating settings (like placement and target sizes)
 */
export const Popover: React.FunctionComponent<FloatingPanelProps> = props => {
    const {
        target,
        placement = 'right',
        strategy = 'absolute',
        children,
        padding = 0,
        constraints,
        middlewares,
        ...otherProps
    } = props

    const floatingReference = useRef<HTMLDivElement>(null)

    useLayoutEffect(() => {
        const floating = floatingReference.current

        if (!floating) {
            return
        }

        function update(): void {
            if (!floating) {
                return
            }

            computePosition(target, floating, {
                placement,
                middleware: middlewares ?? getPositionMiddlewares({ target, floating, strategy, padding, constraints }),
            })
                .then(({ x: xCoordinate, y: yCoordinate, middlewareData }) => {
                    const { referenceHidden } = middlewareData.hide ?? {}

                    Object.assign(floating.style, {
                        position: strategy,
                        top: 0,
                        left: 0,
                        visibility: referenceHidden ? 'hidden' : 'visible',
                        transform: `translate(${Math.round(xCoordinate ?? 0)}px,${Math.round(yCoordinate ?? 0)}px)`,
                    })
                })
                .catch(() => {
                    Object.assign(floating.style, { visibility: 'hidden' })
                })
        }

        // Initial (on mount) tooltip position calculation
        update()

        window.addEventListener('scroll', update, true)

        return () => {
            window.removeEventListener('scroll', update)
        }
    }, [target, floatingReference, placement, strategy, padding, constraints, middlewares])

    return (
        <FloatingPanelContent {...otherProps} portal={strategy === 'fixed'} ref={floatingReference}>
            {children}
        </FloatingPanelContent>
    )
}

interface FloatingPanelContentProps extends React.HTMLAttributes<HTMLDivElement> {
    portal: boolean
}

const FloatingPanelContent = forwardRef<HTMLDivElement, FloatingPanelContentProps>((props, reference) => {
    const { portal, children, ...otherProps } = props
    const element = useRef(portal ? document.createElement('div') : null)

    // Add a container element right after the body tag
    useLayoutEffect(() => {
        const container = element.current
        if (!container) {
            return
        }

        document.body.append(container)

        return () => container.remove()
    }, [])

    return element.current ? (
        createPortal(
            <div ref={reference} {...otherProps}>
                {children}
            </div>,
            element.current
        )
    ) : (
        <div ref={reference} {...otherProps}>
            {children}
        </div>
    )
})
