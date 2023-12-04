import React, { forwardRef, type PropsWithChildren, useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { createPortal } from 'react-dom'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import type { ForwardReferenceComponent } from '../../../../types'
import { createTether, Flipping, Overlapping, type Padding, Position, Strategy, type Tether } from '../../tether'
import type { TetherInstanceAPI } from '../../tether/services/tether-registry'

import styles from './FloatingPanel.module.scss'

const DEFAULT_PADDING: Padding = { top: 8, right: 8, bottom: 8, left: 8 }

export interface FloatingPanelProps extends Omit<Tether, 'target' | 'element'>, React.HTMLAttributes<HTMLDivElement> {
    /**
     * Reference on target HTML element in the DOM.
     * Renders nothing if target isn't specified.
     */
    target: HTMLElement | null

    /**
     * The root element where Popover renders popover content element.
     * This element is used when we render popover with fixed strategy -
     * outside the dom tree.
     */
    rootRender?: HTMLElement | null

    onTetherCreate?: (tether: TetherInstanceAPI) => void
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
        windowPadding = DEFAULT_PADDING,
        constraintPadding,
        targetPadding,
        constraint,
        rootRender,
        onTetherCreate,
        ...otherProps
    } = props

    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)
    const references = useMergeRefs([tooltipReferenceCallback, reference])

    useLayoutEffect(() => {
        if (!tooltipElement) {
            return
        }

        const tether = createTether({
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

        onTetherCreate?.(tether)

        return tether.unsubscribe
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
        onTetherCreate,
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
        rootRender ?? document.body
    )
}) as ForwardReferenceComponent<'div', PropsWithChildren<FloatingPanelProps>>
