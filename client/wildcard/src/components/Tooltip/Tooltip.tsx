import React, { ReactElement } from 'react'

import * as Tooltip from '@radix-ui/react-tooltip'

import styles from './Tooltip.module.scss'

const TooltipProvider: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <Tooltip.Provider delayDuration={0}>{children}</Tooltip.Provider>
)

const TooltipRoot: React.FunctionComponent<React.PropsWithChildren<{}>> = ({ children }) => (
    <Tooltip.Root>{children}</Tooltip.Root>
)

interface TooltipTriggerProps {
    children: ReactElement
}

/**
 * Accepts a single child node. That node will trigger the tooltip content to appear when it is hovered or focused.
 */
const TooltipTrigger: React.FunctionComponent<TooltipTriggerProps> = ({ children }) => (
    <Tooltip.Trigger className={styles.tooltipTrigger}>{children}</Tooltip.Trigger>
)

interface TooltipContentProps {
    children: string
    placement?: Tooltip.TooltipContentProps['side']
}

const TooltipContent: React.FunctionComponent<TooltipContentProps> = ({ children, placement = 'right' }) => (
    <Tooltip.Content side={placement}>
        <Tooltip.Arrow />
        {children}
    </Tooltip.Content>
)

export { TooltipProvider as Provider, TooltipRoot as Root, TooltipTrigger as Trigger, TooltipContent as Content }
