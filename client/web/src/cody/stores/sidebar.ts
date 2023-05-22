import { useCallback } from 'react'

import create from 'zustand'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

import { useIsCodyEnabled } from '../useIsCodyEnabled'

interface CodySidebarSizeStore {
    size: number
    onResize: (size: number) => void
}

const useCodySidebarSizeStore = create<CodySidebarSizeStore>(
    (set): CodySidebarSizeStore => ({
        size: 0,
        onResize(size: number) {
            set({ size })
        },
    })
)

interface CodySidebarOpen {
    isOpen: boolean | undefined
    inputNeedsFocus: boolean
    setFocusProvided: () => void
    setIsOpen: (newValue: boolean | ((previousValue: boolean | undefined) => boolean | undefined) | undefined) => void
}

let inputNeedsFocus = false

// By omitting returning the current size, we don't have to re-render users of this hook (e.g. the
// RepoContainer) on every resize event.
export function useCodySidebarStore(): Omit<CodySidebarSizeStore, 'size'> & CodySidebarOpen {
    const [isOpen, setIsOpen] = useTemporarySetting('cody.showSidebar', false)
    const onResize = useCodySidebarSizeStore(store => store.onResize)

    const setFocusProvided = useCallback(() => {
        inputNeedsFocus = false
    }, [])

    const setSidebarIsOpen = useCallback(
        (...args: Parameters<typeof setIsOpen>) => {
            setIsOpen(...args)
            inputNeedsFocus = true
        },
        [setIsOpen]
    )

    return {
        onResize,
        isOpen,
        inputNeedsFocus,
        setFocusProvided,
        setIsOpen: setSidebarIsOpen,
    }
}

export function useCodySidebarSize(): number {
    const size = useCodySidebarSizeStore(store => store.size)
    const { isOpen } = useCodySidebarStore()
    const enabled = useIsCodyEnabled()

    return isOpen && enabled.sidebar ? size : 0
}
