import create from 'zustand'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary'

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
    setIsOpen: (newValue: boolean | ((previousValue: boolean | undefined) => boolean | undefined) | undefined) => void
}

// By omitting returning the current size, we don't have to re-render users of this hook (e.g. the
// RepoContainer) on every resize event.
export function useCodySidebarStore(): Omit<CodySidebarSizeStore, 'size'> & CodySidebarOpen {
    const [isOpen, setIsOpen] = useTemporarySetting('cody.showSidebar', false)
    const onResize = useCodySidebarSizeStore(store => store.onResize)

    return {
        onResize,
        isOpen,
        setIsOpen,
    }
}

export function useCodySidebarSize(): number {
    const size = useCodySidebarSizeStore(store => store.size)
    const { isOpen } = useCodySidebarStore()
    return isOpen && window.context?.codyEnabled ? size : 0
}
