import React, { forwardRef, PropsWithChildren, useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { createPortal } from 'react-dom'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../../../types'
import { createTether, Flipping, Overlapping, Position, Strategy, Tether } from '../../tether'

import styles from './FloatingPanel.module.scss'

export interface FloatingPanelProps extends Omit<Tether, 'target' | 'element'>, React.HTMLAttributes<HTMLDivElement> {
    /**
     * Reference on target HTML element in the DOM.
     * Renders nothing if target isn't specified.
     */
    target: HTMLElement | null
}

/**
 * React component that wraps up tether positioning logic and provide narrowed down
 * interface of setting to set up floating panel component.
 */
export const FloatingPanel = forwardRef((props, reference) => {
    const {
        as: Component = 'div',
        target,
        marker,
        position = Position.bottomStart,
        overlapping = Overlapping.none,
        flipping = Flipping.all,
        pin = null,
        constrainToScrollParents = true,
        overflowToScrollParents = true,
        strategy = Strategy.Fixed,
        windowPadding,
        constraintPadding,
        targetPadding,
        constraint,
        ...otherProps
    } = props

    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)
    const references = useMergeRefs([tooltipReferenceCallback, reference])

    useLayoutEffect(() => {
        if (!tooltipElement) {
            return
        }

        const { unsubscribe } = createTether({
            element: tooltipElement,
            marker,
            target,
            constraint,
            pin,
            windowPadding,
            constraintPadding,
            targetPadding,
            position,
            strategy,
            overlapping,
            constrainToScrollParents,
            overflowToScrollParents,
            flipping,
        })

        return unsubscribe
    }, [
        target,
        tooltipElement,
        marker,
        constraint,
        windowPadding,
        constraintPadding,
        targetPadding,
        pin,
        position,
        strategy,
        overlapping,
        constrainToScrollParents,
        overflowToScrollParents,
        flipping,
    ])

    if (strategy === Strategy.Absolute) {
        return (
            <Component
                {...otherProps}
                ref={references}
                className={classNames(styles.floatingPanel, styles.floatingPanelAbsolute, otherProps.className)}
            >
                {props.children}
            </Component>
        )
    }

    return createPortal(
        <Component {...otherProps} ref={references} className={classNames(styles.floatingPanel, otherProps.className)}>
            {props.children}
        </Component>,
        document.body
    )
}) as ForwardReferenceComponent<'div', PropsWithChildren<FloatingPanelProps>>
