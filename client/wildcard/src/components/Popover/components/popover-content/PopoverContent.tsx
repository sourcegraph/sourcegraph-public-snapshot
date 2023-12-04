import { forwardRef, useContext, useEffect, useState, type FC, type PropsWithChildren } from 'react'

import classNames from 'classnames'
import FocusLock from 'react-focus-lock'
import { useMergeRefs } from 'use-callback-ref'

import { useKeyboard, useOnClickOutside } from '../../../../hooks'
import type { ForwardReferenceComponent } from '../../../../types'
import { PopoverContext } from '../../contexts/internal-context'
import { PopoverRoot } from '../../contexts/public-context'
import type { TetherInstanceAPI } from '../../tether'
import { type PopoverOpenEvent, PopoverOpenEventReason } from '../../types'
import { FloatingPanel, type FloatingPanelProps } from '../floating-panel/FloatingPanel'

import styles from './PopoverContent.module.scss'

export interface PopoverContentProps extends Omit<FloatingPanelProps, 'target' | 'marker'> {
    isOpen?: boolean
    target?: HTMLElement | null
    tail?: HTMLElement | null
    autoFocus?: boolean
    focusLocked?: boolean
    returnTargetFocus?: boolean
    focusContainerClassName?: string

    onTetherCreate?: (tether: TetherInstanceAPI) => void
    onOpenChange?: (event: PopoverOpenEvent) => void
}

export const PopoverContent = forwardRef(function PopoverContent(props, reference) {
    const {
        children,
        as: Component = 'div',
        target: targetProp,
        tail: tailProp,
        isOpen: isControlledOpen,
        focusLocked = true,
        autoFocus = true,
        returnTargetFocus = true,
        'aria-modal': ariaModal = true,
        role = 'dialog',
        focusContainerClassName,
        onOpenChange: onOpenChangeProp,
        onTetherCreate,
        ...otherProps
    } = props

    const { renderRoot } = useContext(PopoverRoot)

    const {
        anchor,
        isOpen: isOpenContext,
        tailElement: tailContextElement,
        targetElement: contextTargetElement,
        setOpen,
    } = useContext(PopoverContext)

    // Support controlled (higher priority) and context based open state
    const isOpen = isControlledOpen ?? isOpenContext
    const handleOpenChange = onOpenChangeProp ?? setOpen

    const target = targetProp ?? contextTargetElement
    const visualAnchor = anchor?.current ?? target
    const tail = tailProp ?? tailContextElement

    const [popover, setPopover] = useState<HTMLDivElement | null>(null)
    const mergeReference = useMergeRefs([setPopover, reference])

    // Catch any outside click of the popover content element
    useOnClickOutside(mergeReference, event => {
        if (target?.contains(event.target as Node)) {
            return
        }

        handleOpenChange({ isOpen: false, reason: PopoverOpenEventReason.ClickOutside })
    })

    // Close popover on escape
    useKeyboard({ detectKeys: ['Escape'] }, () => {
        // Only fire if we can be sure that the popover is open.
        // Both for controlled and uncontrolled popovers.
        if (isOpen) {
            handleOpenChange({ isOpen: false, reason: PopoverOpenEventReason.Esc })
        }
    })

    if (!isOpen) {
        return null
    }

    return (
        <FloatingPanel
            {...otherProps}
            as={Component}
            ref={mergeReference}
            target={visualAnchor}
            marker={tail}
            role={role}
            aria-modal={ariaModal}
            rootRender={renderRoot}
            className={classNames(styles.popover, otherProps.className)}
            onTetherCreate={onTetherCreate}
        >
            <FloatingPanelContent
                autoFocus={true}
                returnTargetFocus={returnTargetFocus}
                focusLocked={focusLocked}
                popoverElement={popover}
                targetElement={target}
                className={focusContainerClassName}
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
    className?: string
}

const FloatingPanelContent: FC<PropsWithChildren<FloatingPanelContentProps>> = props => {
    const { children, focusLocked, autoFocus, returnTargetFocus, popoverElement, targetElement, className } = props

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
        <FocusLock disabled={!focusLock} returnFocus={returnTargetFocus} className={className}>
            {children}
        </FocusLock>
    )
}
