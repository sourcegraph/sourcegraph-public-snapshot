import { ReactElement } from 'react'

import { TooltipTrigger as SpectrumTooltipRoot, Tooltip as SpectrumTooltipContent } from '@react-spectrum/tooltip'
import { SpectrumTooltipProps } from '@react-types/tooltip'

interface TooltipRootProps {
    children: [ReactElement, ReactElement]
}

const TooltipRoot: React.FunctionComponent<TooltipRootProps> = ({ children }) => (
    <SpectrumTooltipRoot delay={0}>{children}</SpectrumTooltipRoot>
)

interface TooltipTriggerProps {
    children: ReactElement
}

const TooltipTrigger: React.FunctionComponent<TooltipTriggerProps> = ({ children }) => <>{children}</>

interface TooltipContentProps {
    children: string
    position?: SpectrumTooltipProps['placement']
}

const TooltipContent: React.FunctionComponent<TooltipContentProps> = ({ children, position = 'end' }) => (
    <SpectrumTooltipContent placement={position}>{children}</SpectrumTooltipContent>
)

const Root = TooltipRoot
const Trigger = TooltipTrigger
const Content = TooltipContent

export { Root, Trigger, Content }
