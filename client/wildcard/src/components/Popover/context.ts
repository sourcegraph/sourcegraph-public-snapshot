import { createContext, MutableRefObject } from 'react'

import { noop } from 'lodash'

import { PopoverOpenEvent } from './Popover'

export interface PopoverContextData {
    isOpen: boolean
    targetElement: HTMLElement | null
    tailElement: SVGGElement | null
    anchor?: MutableRefObject<HTMLElement | null>
    setOpen: (event: PopoverOpenEvent) => void
    setTargetElement: (element: HTMLElement | null) => void
    setTailElement: (element: SVGGElement | null) => void
}

const DEFAULT_CONTEXT_VALUE: PopoverContextData = {
    isOpen: false,
    targetElement: null,
    tailElement: null,
    setOpen: noop,
    setTargetElement: noop,
    setTailElement: noop,
}

export const PopoverContext = createContext<PopoverContextData>(DEFAULT_CONTEXT_VALUE)
