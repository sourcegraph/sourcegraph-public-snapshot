import React, { FC, forwardRef, ReactElement, useEffect, useRef, useState } from 'react'

import { useId } from '@reach/auto-id'
import { useMergeRefs } from 'use-callback-ref'

import { useDebounce } from '../../hooks'
import { PopoverContent, PopoverOpenEvent, PopoverOpenEventReason, PopoverTail, Position } from '../Popover'

import styles from './Tooltip.module.scss'

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
    content: string | null | undefined

    /** The open state of the tooltip when it is initially rendered. Defaults to `false`. */
    defaultOpen?: boolean

    /**
     * The preferred side of the trigger to render against when open. Will be reversed if
     * a collision is detected. Defaults to `bottom`.
     */
    placement?: `${Position}`
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
    const { children, content, defaultOpen = false, placement = 'bottom' } = props

    const tooltipId = `tooltip-${useId()}`
    const [target, setTarget] = useState<HTMLElement | null>(null)
    const [tail, setTail] = useState<HTMLDivElement | null>(null)
    const popoverContentRef = useRef<HTMLDivElement>(null)

    const [open, setOpen] = useState(defaultOpen)

    useEffect(() => {
        function handleTargetPointerEnter(): void {
            setOpen(true)
        }

        function handleTargetPointerLeave(): void {
            setOpen(false)
        }

        target?.addEventListener('pointerenter', handleTargetPointerEnter)
        target?.addEventListener('pointerleave', handleTargetPointerLeave)
        target?.addEventListener('focus', handleTargetPointerEnter, true)
        target?.addEventListener('blur', handleTargetPointerLeave, true)

        return () => {
            target?.removeEventListener('pointerenter', handleTargetPointerEnter)
            target?.removeEventListener('pointerleave', handleTargetPointerLeave)
            target?.removeEventListener('focus', handleTargetPointerEnter)
            target?.removeEventListener('blur', handleTargetPointerLeave)
        }
    }, [target])

    useEffect(() => {
        const popoverElement = popoverContentRef.current

        function handlePointerEnter(): void {
            setOpen(true)
        }

        function handlePointerLeave(): void {
            setOpen(false)
        }

        popoverElement?.addEventListener('pointerenter', handlePointerEnter)
        popoverElement?.addEventListener('pointerleave', handlePointerLeave)

        return () => {
            popoverElement?.removeEventListener('pointerenter', handlePointerEnter)
            popoverElement?.removeEventListener('pointerleave', handlePointerLeave)
        }
    }, [open])

    const handleOpenChange = (event: PopoverOpenEvent): void => {
        switch (event.reason) {
            case PopoverOpenEventReason.Esc:
            case PopoverOpenEventReason.ClickOutside: {
                setOpen(false)
            }
        }
    }

    const isOpen = useDebounce(open, 100)

    return (
        <>
            <TooltipTarget ref={setTarget} aria-describedby={isOpen ? tooltipId : undefined}>
                {children}
            </TooltipTarget>

            {content && target && isOpen && (
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
    let trigger: React.ReactElement

    // Disabled buttons come through with a disabled prop and must be wrapped with a
    // span in order for the Tooltip to work properly
    // Reference: https://www.radix-ui.com/docs/primitives/components/tooltip#displaying-a-tooltip-from-a-disabled-button
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    if (children.props?.disabled) {
        trigger = (
            <span className={styles.tooltipWrapper}>
                <div className={styles.tooltipTriggerContainer}>
                    {/* eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex */}
                    <div className={styles.tooltipTriggerDisabledOverlay} tabIndex={0} />
                    {children}
                </div>
            </span>
        )
    } else {
        trigger = children
    }

    if (React.isValidElement(trigger)) {
        return React.cloneElement(trigger as ReactElement, {
            'aria-describedby': ariaDescribedby,
            ref: mergedRef,
        })
    }

    return children
})
