import classNames from 'classnames'
import React, { forwardRef, useLayoutEffect, useRef, useState, PropsWithChildren } from 'react'
import { createPortal } from 'react-dom'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../../types'
import { createTether, Flipping, Overlapping, Position, Tether } from '../tether'

import styles from './FloatingPanel.module.scss'

export interface FloatingPanelProps
    extends Omit<Tether, 'target' | 'element' | 'marker'>,
        React.HTMLAttributes<HTMLDivElement> {
    /**
     * Reference on target HTML element in the DOM.
     * Renders nothing if target isn't specified.
     */
    target: HTMLElement | null

    /**
     * Enables tail element rendering and attaches it to
     * floating panel.
     */
    tail?: boolean
}

/**
 * React component that wraps up tether positioning logic and provide narrowed down
 * interface of setting to setup floating panel component.
 */
export const FloatingPanel = forwardRef((props, reference) => {
    const {
        as: Component = 'div',
        target,
        tail,
        position = Position.bottomStart,
        overlapping = Overlapping.none,
        flipping = Flipping.all,
        pin = null,
        constrainToScrollParents = true,
        overflowToScrollParents = true,
        windowPadding,
        constraintPadding,
        constraint,
        ...otherProps
    } = props

    const containerReference = useRef(document.createElement('div'))
    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const [tooltipTailElement, setTooltipTailElement] = useState<HTMLDivElement | null>(null)

    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)

    // Add a container element right after the body tag
    useLayoutEffect(() => {
        const element = containerReference.current

        document.body.append(element)

        return () => {
            element.remove()
        }
    }, [containerReference])

    useLayoutEffect(() => {
        if (!tooltipElement) {
            return
        }

        const { unsubscribe } = createTether({
            element: tooltipElement,
            marker: tooltipTailElement,
            target,
            constraint,
            pin,
            windowPadding,
            constraintPadding,
            position,
            overlapping,
            constrainToScrollParents,
            overflowToScrollParents,
            flipping,
        })

        return unsubscribe
    }, [
        target,
        tooltipElement,
        tooltipTailElement,
        constraint,
        windowPadding,
        constraintPadding,
        pin,
        position,
        overlapping,
        constrainToScrollParents,
        overflowToScrollParents,
        flipping,
    ])

    return createPortal(
        <>
            <Component
                {...otherProps}
                key={position}
                ref={useMergeRefs([tooltipReferenceCallback, reference])}
                className={classNames(styles.floatingPanel, otherProps.className)}
            >
                {props.children}
            </Component>

            {tail && <div className={styles.tail} ref={setTooltipTailElement} />}
        </>,
        containerReference.current
    )
}) as ForwardReferenceComponent<'div', PropsWithChildren<FloatingPanelProps>>
