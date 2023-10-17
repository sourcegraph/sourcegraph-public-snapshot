import React, { type FC, forwardRef, type ReactElement, useCallback, useEffect, useRef, useState } from 'react'

import { useId } from '@reach/auto-id'
import { noop } from 'lodash'
import { useMergeRefs } from 'use-callback-ref'

import { useDebounce } from '../../hooks'
import { PopoverContent, type PopoverOpenEvent, PopoverOpenEventReason, PopoverTail, type Position } from '../Popover'

import styles from './Tooltip.module.scss'

export enum TooltipOpenChangeReason {
    TargetHover = 'TargetHover',
    TargetFocus = 'TargetFocus',
    TargetBlur = 'TargetBlur',
    TargetLeave = 'TargetLeave',
    ClickOutside = 'ClickOutside',
    Esc = 'Esc',
}

export interface TooltipOpenEvent {
    isOpen: boolean
    reason: TooltipOpenChangeReason
}

export interface TooltipProps {
    /**
     * A single child element/component that will trigger the Tooltip to open on hover.
     *
     * **Note:** If you are using a component, it **must** be able to receive and attach
     * a ref (React.forwardRef).
     */
    children: React.ReactElement

    /**
     * The text that will be displayed in the Tooltip. If `null`, no Tooltip will be rendered,
     * allowing for Tooltips to be shown conditionally.
     */
    content: React.ReactNode

    /** The controlled open state prop, it allows to control tooltip appearance from consumer. */
    open?: boolean

    /** The open state of the tooltip when it is initially rendered. Defaults to `false`. */
    defaultOpen?: boolean

    debounce?: number

    /**
     * The preferred side of the trigger to render against when open. Will be reversed if
     * a collision is detected. Defaults to `bottom`.
     */
    placement?: `${Position}`

    /**
     * The open state observer prop. It's supposed to be used with open prop in order to have
     * a fully controlled tooltip.
     */
    onOpenChange?: (event: TooltipOpenEvent) => void
}

/**
 * Renders a Tooltip that will be positioned relative to the wrapped child element. Please
 * reference the examples in Storybook for more details on specific use cases.
 *
 * To support accessibility, our tooltips should:
 * - Be supplemental to the user journey, not essential.
 * - Use clear and concise text.
 * - Not include interactive content (you probably want a `<Popover>` instead).
 *
 * Related accessibility documentation: https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles/tooltip_role
 *
 * In most cases, the child element (trigger) of the Tooltip will not need an `aria-label`
 * attribute, and it should be avoided to prevent repetitive text from being read by a screen
 * reader. However, there are a couple exceptions:
 * - If the trigger is an `<Icon>`, it must have an `aria-label` (and NOT be `aria-hidden`).
 * - If the trigger is a `<Button>` with no visible text within it (e.g., only an icon),
 * it must have an `aria-label`.
 */
export const Tooltip: FC<TooltipProps> = props => {
    const {
        children,
        content,
        open,
        defaultOpen = false,
        placement = 'bottom',
        debounce = 100,
        onOpenChange = noop,
    } = props

    const [target, setTarget] = useState<HTMLElement | null>(null)
    const [tail, setTail] = useState<HTMLDivElement | null>(null)
    const popoverContentRef = useRef<HTMLDivElement>(null)

    const isControlled = open !== undefined
    const [internalOpen, setInternalOpen] = useState(defaultOpen)
    const isOpen = isControlled ? open : internalOpen
    const setOpen = useCallback(
        (event: TooltipOpenEvent): void => {
            if (isControlled) {
                onOpenChange(event)
            } else {
                setInternalOpen(event.isOpen)
            }
        },
        [isControlled, onOpenChange]
    )

    useEffect(() => {
        function handleTargetPointerEnter(): void {
            setOpen({ isOpen: true, reason: TooltipOpenChangeReason.TargetHover })
        }

        function handleTargetPointerLeave(): void {
            setOpen({ isOpen: false, reason: TooltipOpenChangeReason.TargetLeave })
        }

        const preventFocusListeners = shouldPreventFocusListeners(target)

        target?.addEventListener('pointerenter', handleTargetPointerEnter)
        target?.addEventListener('pointerleave', handleTargetPointerLeave)
        if (!preventFocusListeners) {
            target?.addEventListener('focus', handleTargetPointerEnter, true)
            target?.addEventListener('blur', handleTargetPointerLeave, true)
        }

        return () => {
            target?.removeEventListener('pointerenter', handleTargetPointerEnter)
            target?.removeEventListener('pointerleave', handleTargetPointerLeave)
            if (!preventFocusListeners) {
                target?.removeEventListener('focus', handleTargetPointerEnter)
                target?.removeEventListener('blur', handleTargetPointerLeave)
            }
        }
    }, [target, setOpen])

    useEffect(() => {
        const popoverElement = popoverContentRef.current

        function handlePointerEnter(): void {
            setOpen({ isOpen: true, reason: TooltipOpenChangeReason.TargetHover })
        }

        function handlePointerLeave(): void {
            setOpen({ isOpen: false, reason: TooltipOpenChangeReason.TargetLeave })
        }

        popoverElement?.addEventListener('pointerenter', handlePointerEnter)
        popoverElement?.addEventListener('pointerleave', handlePointerLeave)

        return () => {
            popoverElement?.removeEventListener('pointerenter', handlePointerEnter)
            popoverElement?.removeEventListener('pointerleave', handlePointerLeave)
        }
    }, [isOpen, setOpen])

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        switch (event.reason) {
            case PopoverOpenEventReason.Esc: {
                setOpen({ isOpen: event.isOpen, reason: TooltipOpenChangeReason.Esc })
                return
            }
            case PopoverOpenEventReason.ClickOutside: {
                setOpen({ isOpen: event.isOpen, reason: TooltipOpenChangeReason.ClickOutside })
                return
            }
        }
    }

    const tooltipId = `tooltip-${useId()}`
    const isOpenDebounced = useDebounce(isOpen, debounce)

    return (
        <>
            <TooltipTarget ref={setTarget} aria-describedby={isOpenDebounced ? tooltipId : undefined}>
                {children}
            </TooltipTarget>

            {content && target && isOpenDebounced && (
                <>
                    <PopoverContent
                        role="tooltip"
                        id={tooltipId}
                        ref={popoverContentRef}
                        isOpen={true}
                        target={target}
                        tail={tail}
                        position={placement}
                        focusLocked={false}
                        autoFocus={false}
                        returnTargetFocus={false}
                        className={styles.tooltipContent}
                        onOpenChange={handleOpenChange}
                    >
                        {content}
                    </PopoverContent>

                    <PopoverTail ref={setTail} forceRender={true} size="sm" className={styles.tooltipArrow} />
                </>
            )}
        </>
    )
}

interface TooltipTargetProps {
    'aria-describedby'?: string | undefined
    children: React.ReactElement
}

const TooltipTarget = forwardRef<any, TooltipTargetProps>(function TooltipTarget(props, forwardedRef) {
    const { 'aria-describedby': ariaDescribedby, children } = props

    const mergedRef = useMergeRefs([forwardedRef, (children as any).ref])

    if (React.isValidElement(children)) {
        return React.cloneElement(children as ReactElement, {
            'aria-describedby': ariaDescribedby,
            ref: mergedRef,
        })
    }

    return children
})

// We use this test to work around a Chromium bug that causes an `<svg>` element with a focus event
// listener to appear in the tab-order. See https://bugs.chromium.org/p/chromium/issues/detail?id=445798
function shouldPreventFocusListeners(target: HTMLElement | null): boolean {
    return (
        target?.tagName === 'svg' &&
        (target.getAttribute('tabindex') === '-1' || target?.getAttribute('tabindex') === null)
    )
}
