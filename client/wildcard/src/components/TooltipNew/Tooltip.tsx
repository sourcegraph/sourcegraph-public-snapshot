import React from 'react'

import {
    Root,
    Trigger,
    Content,
    Provider,
    TooltipProps as PrimitiveTooltipProps,
    TooltipContentProps as PrimitiveTooltipContentProps,
} from '@radix-ui/react-tooltip'

export interface TooltipProps extends PrimitiveTooltipProps, PrimitiveTooltipContentProps {
    title: string
}

export const Tooltip: React.FunctionComponent<TooltipProps> = ({ children, title, ...props }) => {
    const [open, setOpen] = React.useState(false)
    const isPointerDownOnContentReference = React.useRef(false)

    React.useEffect(() => {
        const handlePointerUp = () => (isPointerDownOnContentReference.current = false)
        document.addEventListener('pointerup', handlePointerUp)
        return () => document.removeEventListener('pointerup', handlePointerUp)
    }, [])

    return (
        <Root
            open={open}
            onOpenChange={open => {
                if (open) {
                    setOpen(true)
                } else if (!isPointerDownOnContentReference.current) {
                    setOpen(false)
                }
            }}
        >
            <Trigger asChild={true}>
                {/* eslint-disable-next-line react/forbid-dom-props */}
                <div style={{ display: 'inherit' }}>
                    {children}
                    <Content
                        align="start"
                        onPointerDown={() => (isPointerDownOnContentReference.current = true)}
                        style={{
                            background: 'black',
                            color: 'white',
                            borderRadius: 4,
                            padding: '0.375rem',
                            marginTop: '0.25rem',
                        }}
                        {...props}
                    >
                        {title}
                    </Content>
                </div>
            </Trigger>
        </Root>
    )
}

export const TooltipProvider = Provider
