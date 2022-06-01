import create, { GetState, SetState } from 'zustand'
import { persist, StoreApiWithPersist } from 'zustand/middleware'

import { TourLanguage } from '../tour/components/Tour/types'

interface QuickStartTourState {
    completedStepIds?: string[]
    status?: 'closed' | 'completed'
    language?: TourLanguage
}

export interface QuickStartTourListState {
    tours: Record<string, QuickStartTourState>
    setCompletedStepIds: (key: string, id: string[]) => void
    setLanguage: (key: string, id: TourLanguage) => void
    setStatus: (key: string, status: QuickStartTourState['status']) => void
    resetTour: (key: string) => void
}

export const useQuickStartTourListState = create<QuickStartTourListState>(
    persist<
        QuickStartTourListState,
        SetState<QuickStartTourListState>,
        GetState<QuickStartTourListState>,
        StoreApiWithPersist<QuickStartTourListState>
    >(
        (set, get): QuickStartTourListState => ({
            tours: {},
            setLanguage: (key, language) =>
                set({ tours: { ...get().tours, [key]: { ...get().tours[key], language } } }),
            setCompletedStepIds: (key, completedStepIds) =>
                set({ tours: { ...get().tours, [key]: { ...get().tours[key], completedStepIds } } }),
            setStatus: (key, status) => set({ tours: { ...get().tours, [key]: { ...get().tours[key], status } } }),
            resetTour: (key: string) => set({ tours: { ...get().tours, [key]: {} } }),
        }),
        {
            name: 'quick-start-tour',
        }
    )
)
