import React from 'react'

import * as TooltipPrimitive from '@radix-ui/react-tooltip'
import { isEmpty } from 'lodash'

import styles from './Tooltip.module.scss'

export interface TooltipProps {
    /**
     * A single child element/component that will trigger the Tooltip to open on hover.
     *
     * **Note:** If you are using a component, it **must** be able to receive and attach a ref (React.forwardRef).
     **/
    children: React.ReactElement
    /** The text that will be displayed in the Tooltip. If `null`, no Tooltip will be rendered, allowing for Tooltips to be shown conditionally. */
    content: string | null | undefined
    /** The open state of the tooltip when it is initially rendered. Defaults to `false`. */
    defaultOpen?: boolean
    /** The preferred side of the trigger to render against when open. Will be reversed if a collision is detected. Defaults to `bottom`. */
    placement?: TooltipPrimitive.TooltipContentProps['side']
}

/** Arrow width in pixels */
const TOOLTIP_ARROW_WIDTH = 14
/** Arrow height in pixel */
const TOOLTIP_ARROW_HEIGHT = 6

// Handling the onPointerDownOutside event and preventing the default behavior allows us to keep the Tooltip content open
// even if the trigger <span> was clicked; this allows buttons to be clicked and text to be selected without dismissing content.
// Reference: https://github.com/radix-ui/primitives/issues/1077
function onPointerDownOutside(event: Event): void {
    event.preventDefault()
}

/**
 * Renders a Tooltip that will be positioned relative to the wrapped child element. Please reference the examples in Storybook
 * for more details on specific use cases.
 *
 * **NOTE:** The Tooltip implementation currently breaks the behavior of triggers that use `ButtonLink` with no `to` prop. Specifically,
 * the onClick handler of `<ButtonLink>` does not get composed correctly, and the default behavior will not be prevented when that component
 * has an empty href (resulting in a page reload). If the trigger element you are using is not working as expected, please wrap that
 * element with an additional element (such as a `<span>`). That should resolve the issue.
 *
 * To support accessibility, our tooltips should:
 * - Be supplemental to the user journey, not essential.
 * - Use clear and concise text.
 * - Not include interactive content (you probably want a `<Popover>` instead).
 *
 * Related accessibility documentation: https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles/tooltip_role
 *
 * In most cases, the child element (trigger) of the Tooltip will not need an `aria-label` attribute, and it should be avoided
 * to prevent repetitive text from being read by a screen reader. However, there are a couple exceptions:
 * - If the trigger is an `<Icon>`, it must have an `aria-label` (and NOT be `aria-hidden`).
 * - If the trigger is a `<Button>` with no visible text within it (e.g., only an icon), it must have an `aria-label`.
 *
 * To test for the correct content in test suites where the tooltip won't be opened, please use `data-*` attributes on the trigger element.
 */
export const Tooltip: React.FunctionComponent<TooltipProps> = ({
    children,
    content,
    defaultOpen = false,
    placement = 'bottom',
}) => {
    let trigger: React.ReactElement
    // Disabled buttons come through with a disabled prop and must be wrapped with a span in order for the Tooltip to work properly
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

    // NOTE: We plan to consolidate this logic with our Popover component in the future, but chose Radix first to support short-term accessibility needs.
    // GitHub issue: https://github.com/sourcegraph/sourcegraph/issues/36080
    return (
        // The small delayDuration helps prevent the tooltip from immediately closing when it gets triggered in the
        // exact spot the arrow is overlapping the content (allows time for the cursor to move more naturally)
        <TooltipPrimitive.Root delayDuration={100} defaultOpen={defaultOpen}>
            <TooltipPrimitive.Trigger asChild={true}>{trigger}</TooltipPrimitive.Trigger>
            {
                // The rest of the Tooltip components still need to be rendered for the content to correctly be shown conditionally.
                isEmpty(content) ? null : (
                    <TooltipPrimitive.TooltipContent
                        onPointerDownOutside={onPointerDownOutside}
                        className={styles.tooltipContent}
                        side={placement}
                        role="tooltip"
                        // This offset helps prevent the tooltip from immediately closing when it gets triggered in the
                        // exact spot the arrow is overlapping the content
                        alignOffset={1}
                    >
                        {content}

                        <TooltipPrimitive.Arrow
                            className={styles.tooltipArrow}
                            height={TOOLTIP_ARROW_HEIGHT}
                            width={TOOLTIP_ARROW_WIDTH}
                        />
                    </TooltipPrimitive.TooltipContent>
                )
            }
        </TooltipPrimitive.Root>
    )
}
