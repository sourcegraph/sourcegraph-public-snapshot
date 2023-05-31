import create from 'zustand'

export const useSidebarSize = create<{ sidebarSize: number; setSidebarSize: (size: number) => void }>(set => ({
    sidebarSize: 0,
    setSidebarSize: (size: number) => set({ sidebarSize: size }),
}))
