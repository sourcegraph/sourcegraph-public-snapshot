import React, { useState } from 'react'

import { ForwardReferenceComponent } from '../../types'
import { Popover, PopoverContent, PopoverTrigger, Position } from '../Popover'

function composeEventHandlers<E>(
    originalEventHandler?: (event: E) => void,
    ourEventHandler?: (event: E) => void,
    { checkForDefaultPrevented = true } = {}
) {
    return function handleEvent(event: E) {
        originalEventHandler?.(event)

        if (checkForDefaultPrevented === false || !((event as unknown) as Event).defaultPrevented) {
            return ourEventHandler?.(event)
        }
    }
}

/**
 * <Tooltip title="Hello world">
 *  <Button>Test</Button>
 * </Tooltip>
 *
 */

interface TooltipRootProps {}
export const TooltipRoot: React.FunctionComponent<TooltipRootProps> = ({ children }) => <>{children}</>

interface TooltipTriggerProps {
    asChild?: boolean
    context: any
}

const TooltipTrigger = React.forwardRef((props, reference) => {
    const { context, ...triggerProps } = props
    const isPointerDownReference = React.useRef(false)
    const handlePointerUp = React.useCallback(() => (isPointerDownReference.current = false), [])
    React.useEffect(() => () => document.removeEventListener('pointerup', handlePointerUp), [handlePointerUp])

    return (
        <PopoverTrigger
            asChild={true}
            // // We purposefully avoid adding `type=button` here because tooltip triggers are also
            // // commonly anchors and the anchor `type` attribute signifies MIME type.
            // aria-describedby={context.open ? context.contentId : undefined}
            // data-state={context.stateAttribute}
            {...triggerProps}
            ref={reference}
            onPointerEnter={composeEventHandlers(props.onPointerEnter, event => {
                if (event.pointerType !== 'touch') {
                    context.onTriggerEnter()
                }
            })}
            onPointerLeave={composeEventHandlers(props.onPointerLeave, context.onClose)}
            onPointerDown={composeEventHandlers(props.onPointerDown, () => {
                isPointerDownReference.current = true
                document.addEventListener('pointerup', handlePointerUp, { once: true })
            })}
            onFocus={composeEventHandlers(props.onFocus, () => {
                if (!isPointerDownReference.current) {
                    context.onOpen()
                }
            })}
            onBlur={composeEventHandlers(props.onBlur, context.onClose)}
            onClick={composeEventHandlers(props.onClick, event => {
                // keyboard click will occur under different conditions for different node
                // types so we use `onClick` instead of `onKeyDown` to respect that
                const isKeyboardClick = event.detail === 0
                if (isKeyboardClick) {
                    context.onClose()
                }
            })}
        />
    )
}) as ForwardReferenceComponent<'button', TooltipTriggerProps>

export interface TooltipProps {
    title: string
}
export const Tooltip: React.FunctionComponent<TooltipProps> = ({ children, title }) => {
    const [open, setOpen] = useState(false)
    console.log('open state', open)
    const context = {
        onTriggerEnter: () => {
            console.log('trigger enter')
            setOpen(true)
        },
        onClose: () => {
            console.log('close')
            setOpen(false)
        },
        onOpen: () => {
            console.log('open')
            setOpen(true)
        },
    }

    return (
        <Popover isOpen={open} onOpenChange={event => setOpen(event.isOpen)}>
            <TooltipRoot>
                <TooltipTrigger context={context}>{children}</TooltipTrigger>
                <PopoverContent position={Position.top} tail={true}>
                    {title}
                </PopoverContent>
            </TooltipRoot>
        </Popover>
    )
}
