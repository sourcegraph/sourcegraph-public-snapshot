import {
    type FunctionComponent,
    type MutableRefObject,
    type PropsWithChildren,
    useCallback,
    useContext,
    useMemo,
    useState,
} from 'react'

import { noop } from 'lodash'

import { PopoverContext } from './contexts/internal-context'
import type { PopoverOpenEvent } from './types'

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
    const [tailElement, setTailElement] = useState<HTMLElement | null>(null)

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

interface usePopoverContextData {
    isOpen: boolean
}

/**
 * Public entry point for getting information about popover state.
 * Note: that this hook shouldn't expose any set-like internal state
 * methods.
 */
export function usePopoverContext(): usePopoverContextData {
    const { isOpen } = useContext(PopoverContext)

    return { isOpen }
}
