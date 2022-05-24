import React, { ReactNode } from 'react'

import * as TooltipPrimitive from '@radix-ui/react-tooltip'

import styles from './Tooltip.module.scss'

interface TooltipProps {
    /** A single child element that will trigger the Tooltip to open on hover. */
    children: ReactNode
    /** The text that will be displayed in the Tooltip. */
    content: string
    /** The open state of the tooltip when it is initially rendered. */
    defaultOpen?: boolean
    /** The preferred side of the trigger to render against when open. Will be reversed if a collision is detected. */
    placement?: TooltipPrimitive.TooltipContentProps['side']
}

/** Arrow width in pixels */
const TOOLTIP_ARROW_WIDTH = 14
/** Arrow height in pixel */
const TOOLTIP_ARROW_HEIGHT = 6

export const Tooltip: React.FunctionComponent<TooltipProps> = ({
    children,
    content,
    defaultOpen = false,
    placement = 'right',
}) => (
    <TooltipPrimitive.Root delayDuration={0} defaultOpen={defaultOpen}>
        <TooltipPrimitive.Trigger asChild={true}>
            <span className={styles.tooltip}>
                {children}

                {/*
                 * Rendering the Content within the Trigger is a workaround to support being able to hover over the Tooltip content itself.
                 * Refrence: https://github.com/radix-ui/primitives/issues/620#issuecomment-1079147761
                 */}
                <TooltipPrimitive.TooltipContent className={styles.tooltipContent} side={placement}>
                    {content}

                    <TooltipPrimitive.Arrow
                        className={styles.tooltipArrow}
                        height={TOOLTIP_ARROW_HEIGHT}
                        width={TOOLTIP_ARROW_WIDTH}
                    />
                </TooltipPrimitive.TooltipContent>
            </span>
        </TooltipPrimitive.Trigger>
    </TooltipPrimitive.Root>
)
