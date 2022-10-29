import { forwardRef, useContext, useEffect, useState, FC, PropsWithChildren } from 'react'

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
    returnTargetFocus?: boolean
}

export const PopoverContent = forwardRef(function PopoverContent(props, reference) {
    const {
        isOpen,
        children,
        targetElement: propertyTargetElement = null,
        focusLocked = true,
        autoFocus = true,
        returnTargetFocus = true,
        as: Component = 'div',
        role = 'dialog',
        'aria-modal': ariaModal = true,
        ...otherProps
    } = props

    const { renderRoot } = useContext(PopoverRoot)

    const { isOpen: isOpenContext, targetElement: contextTargetElement, tailElement, anchor, setOpen } = useContext(
        PopoverContext
    )

    const targetElement = contextTargetElement ?? propertyTargetElement
    const anchorOrTargetElement = anchor?.current ?? targetElement
    const [tooltipElement, setTooltipElement] = useState<HTMLDivElement | null>(null)
    const tooltipReferenceCallback = useCallbackRef<HTMLDivElement>(null, setTooltipElement)
    const mergeReference = useMergeRefs([tooltipReferenceCallback, reference])

    // Catch any outside click of the popover content element
    useOnClickOutside(mergeReference, event => {
        if (anchorOrTargetElement?.contains(event.target as Node)) {
            return
        }

        setOpen({ isOpen: false, reason: PopoverOpenEventReason.ClickOutside })
    })

    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, () => {
        // Only fire if we can be sure that the popover is open.
        // Both for controlled and uncontrolled popovers.
        if (isOpen || isOpenContext) {
            setOpen({ isOpen: false, reason: PopoverOpenEventReason.Esc })
        }
    })

    if (!isOpenContext && !isOpen) {
        return null
    }

    return (
        <FloatingPanel
            {...otherProps}
            as={Component}
            ref={mergeReference}
            target={anchorOrTargetElement}
            marker={tailElement}
            role={role}
            aria-modal={ariaModal}
            rootRender={renderRoot}
            className={classNames(styles.popover, otherProps.className)}
        >
            <FloatingPanelContent
                autoFocus={true}
                returnTargetFocus={returnTargetFocus}
                focusLocked={focusLocked}
                popoverElement={tooltipElement}
                targetElement={targetElement}
            >
                {children}
            </FloatingPanelContent>
        </FloatingPanel>
    )
}) as ForwardReferenceComponent<'div', PopoverContentProps>

interface FloatingPanelContentProps {
    focusLocked: boolean
    autoFocus: boolean
    returnTargetFocus: boolean
    popoverElement: HTMLElement | null
    targetElement: HTMLElement | null
}

const FloatingPanelContent: FC<PropsWithChildren<FloatingPanelContentProps>> = props => {
    const { children, focusLocked, autoFocus, returnTargetFocus, popoverElement, targetElement } = props

    const [focusLock, setFocusLock] = useState(false)

    // Native behavior of browsers about focus elements says - if element that gets focus
    // is in outside the visible viewport then browser should scroll to this element automatically.
    // This logic breaks popover behavior by loosing scroll positions of the scroll container with
    // target element. In order to preserve scroll we should adjust order of actions
    // Render popover element in the DOM → Calculate and apply the right position for the popover →
    // Enable focus lock (therefore autofocus first scrollable element within the popover content)
    useEffect(() => {
        if (popoverElement && autoFocus && focusLocked) {
            requestAnimationFrame(() => {
                setFocusLock(true)
            })
        }

        return () => setFocusLock(false)
    }, [autoFocus, focusLocked, popoverElement])

    // In some cases FocusLock doesn't return focus to the popover trigger (target) element
    // In order to ensure that return focus logic works always we explicitly do this manually
    // in the following hook.
    useEffect(
        () => () => {
            if (returnTargetFocus) {
                targetElement?.focus({ preventScroll: true })
            }
        },
        [targetElement, returnTargetFocus]
    )

    if (!focusLocked) {
        return <>{children}</>
    }

    return (
        <FocusLock disabled={!focusLock} returnFocus={returnTargetFocus}>
            {children}
        </FocusLock>
    )
}
