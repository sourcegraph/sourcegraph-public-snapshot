import React, { forwardRef, useContext } from 'react'

import { noop } from 'lodash'
import { useCallbackRef, useMergeRefs } from 'use-callback-ref'

import { ForwardReferenceComponent } from '../../../types'
import { PopoverContext } from '../contexts/internal-context'
import { PopoverOpenEventReason } from '../Popover'

interface PopoverTriggerProps {}

export const PopoverTrigger = forwardRef(function PopoverTrigger(props, reference) {
    const { as: Component = 'button', onClick = noop, ...otherProps } = props
    const { setTargetElement, setOpen, isOpen } = useContext(PopoverContext)

    const callbackReference = useCallbackRef<HTMLButtonElement>(null, setTargetElement)
    const mergedReference = useMergeRefs([reference, callbackReference])

    const handleClick: React.MouseEventHandler<HTMLButtonElement> = event => {
        setOpen({ isOpen: !isOpen, reason: PopoverOpenEventReason.TriggerClick })
        onClick(event)
    }

    return <Component ref={mergedReference} onClick={handleClick} {...otherProps} />
}) as ForwardReferenceComponent<'button', PopoverTriggerProps>
