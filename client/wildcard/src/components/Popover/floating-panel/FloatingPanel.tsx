import React, { forwardRef, PropsWithChildren, useLayoutEffect, useRef, useState } from 'react'

import classNames from 'classnames'
import { createPortal } from 'react-dom'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../../types'
import { createTether, Flipping, Overlapping, Position, Strategy, Tether } from '../tether'

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

    /**
     * Class name for the tail element
     */
    tailClassName?: string
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
        strategy = Strategy.Fixed,
        windowPadding,
        constraintPadding,
        targetPadding,
        constraint,
        tailClassName,
        ...otherProps
    } = props

    const containerReference = useRef(document.createElement('div'))
    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const [tooltipTailElement, setTooltipTailElement] = useState<HTMLDivElement | null>(null)
    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)
    const references = useMergeRefs([tooltipReferenceCallback, reference])

    // Add a container element right after the body tag
    useLayoutEffect(() => {
        if (strategy === Strategy.Absolute) {
            return
        }

        const element = containerReference.current

        document.body.append(element)

        return () => {
            element.remove()
        }
    }, [containerReference, strategy])

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
        tooltipTailElement,
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

    const tailClassNames = tail
        ? classNames(styles.tail, tailClassName, { [styles.tailAbsolute]: strategy === Strategy.Absolute })
        : undefined

    if (strategy === Strategy.Absolute) {
        return (
            <>
                <Component
                    {...otherProps}
                    ref={references}
                    className={classNames(styles.floatingPanel, styles.floatingPanelAbsolute, otherProps.className)}
                >
                    {props.children}
                </Component>

                {tail && <div className={tailClassNames} ref={setTooltipTailElement} />}
            </>
        )
    }

    return createPortal(
        <>
            <Component
                {...otherProps}
                ref={references}
                className={classNames(styles.floatingPanel, otherProps.className)}
            >
                {props.children}
            </Component>

            {tail && <div className={tailClassNames} ref={setTooltipTailElement} />}
        </>,
        containerReference.current
    )
}) as ForwardReferenceComponent<'div', PropsWithChildren<FloatingPanelProps>>
