import { forwardRef, useContext, useEffect, useState } from 'react'

import classNames from 'classnames'
import FocusLock from 'react-focus-lock'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import { useKeyboard, useOnClickOutside } from '../../../../hooks'
import { ForwardReferenceComponent } from '../../../../types'
import { PopoverContext } from '../../contexts/internal-context'
import { PopoverRoot } from '../../contexts/public-context'
import { PopoverOpenEventReason } from '../../Popover'
import { FloatingPanel, FloatingPanelProps } from '../floating-panel/FloatingPanel'

import styles from './PopoverContent.module.scss'

export interface PopoverContentProps extends Omit<FloatingPanelProps, 'target' | 'marker'> {
    isOpen?: boolean
    focusLocked?: boolean
    autoFocus?: boolean
    targetElement?: HTMLElement | null
}

export const PopoverContent = forwardRef(function PopoverContent(props, reference) {
    const {
        isOpen,
        children,
        targetElement: propertyTargetElement = null,
        focusLocked = true,
        autoFocus = true,
        as: Component = 'div',
        role = 'dialog',
        'aria-modal': ariaModel = true,
        ...otherProps
    } = props

    const { isOpen: isOpenContext, targetElement: contextTargetElement, tailElement, anchor, setOpen } = useContext(
        PopoverContext
    )
    const { renderRoot } = useContext(PopoverRoot)

    const targetElement = contextTargetElement ?? propertyTargetElement
    const [focusLock, setFocusLock] = useState(false)
    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)
    const mergeReference = useMergeRefs([tooltipReferenceCallback, reference])

    // Catch any outside click of the popover content element
    useOnClickOutside(mergeReference, event => {
        if (targetElement?.contains(event.target as Node)) {
            return
        }

        setOpen({ isOpen: false, reason: PopoverOpenEventReason.ClickOutside })
    })

    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, () => setOpen({ isOpen: false, reason: PopoverOpenEventReason.Esc }))

    // Native behavior of browsers about focus elements says - if element that gets focus
    // is in outside the visible viewport then browser should scroll to this element automatically.
    // This logic breaks popover behavior by loosing scroll positions of the scroll container with
    // target element. In order to preserve scroll we should adjust order of actions
    // Render popover element in the DOM → Calculate and apply the right position for the popover →
    // Enable focus lock (therefore autofocus first scrollable element within the popover content)
    useEffect(() => {
        if (tooltipElement && autoFocus && focusLocked) {
            requestAnimationFrame(() => {
                setFocusLock(true)
            })
        }

        return () => setFocusLock(false)
    }, [autoFocus, focusLocked, tooltipElement])

    if (!isOpenContext && !isOpen) {
        return null
    }

    return (
        <FloatingPanel
            {...otherProps}
            as={Component}
            ref={mergeReference}
            target={anchor?.current ?? targetElement}
            marker={tailElement}
            role={role}
            aria-modal={ariaModel}
            rootRender={renderRoot}
            className={classNames(styles.popover, otherProps.className)}
        >
            {focusLocked ? (
                <FocusLock disabled={!focusLock} returnFocus={true}>
                    {children}
                </FocusLock>
            ) : (
                children
            )}
        </FloatingPanel>
    )
}) as ForwardReferenceComponent<'div', PopoverContentProps>
