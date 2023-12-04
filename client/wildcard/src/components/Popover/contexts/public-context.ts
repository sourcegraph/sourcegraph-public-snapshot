import { createContext } from 'react'

interface PopoverRootData {
    renderRoot: HTMLElement | null
}

const DEFAULT_POPOVER_PROVIDER_INFO: PopoverRootData = {
    renderRoot: null,
}

export const PopoverRoot = createContext<PopoverRootData>(DEFAULT_POPOVER_PROVIDER_INFO)
