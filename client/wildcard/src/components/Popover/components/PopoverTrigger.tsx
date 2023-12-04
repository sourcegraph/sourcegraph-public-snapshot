import React, { forwardRef, type ReactNode, useContext } from 'react'

import { noop } from 'lodash'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import type { ForwardReferenceComponent } from '../../../types'
import { PopoverContext } from '../contexts/internal-context'
import { PopoverOpenEventReason } from '../types'

export interface PopoverTriggerProps {
    children?: ReactNode | ((isOpen: boolean) => ReactNode)
}

export const PopoverTrigger = forwardRef(function PopoverTrigger(props, reference) {
    const { as: Component = 'button', onClick = noop, children, ...otherProps } = props
    const { setTargetElement, setOpen, isOpen } = useContext(PopoverContext)

    const callbackReference = useCallbackRef<HTMLButtonElement>(null, setTargetElement)
    const mergedReference = useMergeRefs([reference, callbackReference])

    const handleClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        setOpen({ isOpen: !isOpen, reason: PopoverOpenEventReason.TriggerClick })
        onClick(event)
    }

    return (
        <Component ref={mergedReference} aria-expanded={isOpen} onClick={handleClick} {...otherProps}>
            {typeof children === 'function' ? children(isOpen) : children}
        </Component>
    )
}) as ForwardReferenceComponent<'button', PopoverTriggerProps>
