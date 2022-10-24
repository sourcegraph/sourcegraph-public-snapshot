import React, { FC, forwardRef, ReactElement, useEffect, useRef, useState } from 'react'

import classNames from 'classnames'
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
 * **NOTE:** The Tooltip implementation currently breaks the behavior of triggers that use
 * `ButtonLink` with no `to` prop. Specifically, the onClick handler of `<ButtonLink>` does
 * not get composed correctly, and the default behavior will not be prevented when that
 * component has an empty href (resulting in a page reload). If the trigger element you are
 * using is not working as expected, please wrap that element with an additional element
 * (such as a `<span>`). That should resolve the issue.
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
 *
 * To test for the correct content in test suites where the tooltip won't be opened, please
 * use `data-*` attributes on the trigger element.
 */
export const Tooltip: FC<TooltipProps> = props => {
    const { children, content, defaultOpen = false, placement = 'bottom' } = props

    const [target, setTarget] = useState<HTMLElement | null>(null)
    const [tail, setTail] = useState<HTMLDivElement | null>(null)
    const popoverContentRef = useRef<HTMLDivElement>(null)

    const [open, setOpen] = useState(defaultOpen)

    useEffect(() => {
        function handlePointerOver(): void {
            setOpen(true)
        }

        function handlePointerLeave(): void {
            setOpen(false)
        }

        target?.addEventListener('pointerover', handlePointerOver)
        target?.addEventListener('pointerleave', handlePointerLeave)
        target?.addEventListener('focus', handlePointerOver, true)
        target?.addEventListener('blur', handlePointerLeave, true)

        return () => {
            target?.removeEventListener('pointerenter', handlePointerOver)
            target?.removeEventListener('pointerleave', handlePointerLeave)
            target?.removeEventListener('focus', handlePointerOver)
            target?.removeEventListener('blur', handlePointerLeave)
        }
    }, [target])

    useEffect(() => {
        const popoverElement = popoverContentRef.current

        function handlePointerOver(): void {
            setOpen(true)
        }

        function handlePointerLeave(): void {
            setOpen(false)
        }

        popoverElement?.addEventListener('pointerenter', handlePointerOver)
        popoverElement?.addEventListener('pointerleave', handlePointerLeave)

        return () => {
            popoverElement?.removeEventListener('pointerenter', handlePointerOver)
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
            <TooltipTarget ref={setTarget}>{children}</TooltipTarget>

            {content && target && (
                <>
                    <PopoverContent
                        role="tooltip"
                        ref={popoverContentRef}
                        isOpen={isOpen}
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

                    {isOpen && <PopoverTail ref={setTail} forceRender={true} className={styles.tooltipArrow} />}
                </>
            )}
        </>
    )
}

interface TooltipTargetProps {
    children: React.ReactElement
}

const TooltipTarget = forwardRef<any, TooltipTargetProps>((props, forwardedRef) => {
    const { children } = props

    const mergedRef = useMergeRefs([forwardedRef, (children as any).ref])
    let trigger: React.ReactElement

    // Disabled buttons come through with a disabled prop and must be wrapped with a
    // span in order for the Tooltip to work properly
    // Reference: https://www.radix-ui.com/docs/primitives/components/tooltip#displaying-a-tooltip-from-a-disabled-button
    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access
    if (children.props?.disabled) {
        trigger = (
            <span className={classNames(styles.tooltipWrapper, children.props?.className)}>
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
            ref: mergedRef,
        })
    }

    return children
})
