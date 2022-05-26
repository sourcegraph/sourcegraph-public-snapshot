import { FunctionComponent, MutableRefObject, PropsWithChildren, useCallback, useMemo, useState } from 'react'

import { noop } from 'lodash'

import { PopoverContext } from './context'

export enum PopoverOpenEventReason {
    TriggerClick = 'TriggerClick',
    TriggerFocus = 'TriggerFocus',
    TriggerBlur = 'TriggerBlur',
    ClickOutside = 'ClickOutside',
    Esc = 'Esc',
}

export interface PopoverOpenEvent {
    isOpen: boolean
    reason: PopoverOpenEventReason
}

type PopoverControlledProps =
    | { isOpen?: undefined; onOpenChange?: never }
    | { isOpen: boolean; onOpenChange: (event: PopoverOpenEvent) => void }

interface PopoverCommonProps {
    anchor?: MutableRefObject<HTMLElement | null>
}

export type PopoverProps = PopoverCommonProps & PopoverControlledProps

/**
 * Returns a root component for the compound popover components family.
 * Renders nothing but gathers all vital compound context information.
 */
export const Popover: FunctionComponent<PropsWithChildren<PopoverProps>> = props => {
    const { children, anchor, isOpen, onOpenChange = noop } = props

    const [targetElement, setTargetElement] = useState<HTMLElement | null>(null)
    const [tailElement, setTailElement] = useState<SVGGElement | null>(null)

    const [isInternalOpen, setInternalOpen] = useState<boolean>(false)
    const isControlled = isOpen !== undefined
    const isPopoverOpen = isControlled ? isOpen : isInternalOpen
    const setOpen = useCallback<(event: PopoverOpenEvent) => void>(
        event => {
            if (!isControlled) {
                setInternalOpen(event.isOpen)
                return
            }

            onOpenChange(event)
        },
        [isControlled, onOpenChange]
    )

    const context = useMemo(
        () => ({
            isOpen: isPopoverOpen,
            targetElement,
            tailElement,
            anchor,
            setOpen,
            setTargetElement,
            setTailElement,
        }),
        [isPopoverOpen, targetElement, tailElement, anchor, setOpen]
    )

    return <PopoverContext.Provider value={context}>{children}</PopoverContext.Provider>
}
