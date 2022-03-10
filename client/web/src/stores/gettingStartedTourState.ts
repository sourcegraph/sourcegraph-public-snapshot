import create, { GetState, SetState } from 'zustand'
import { persist, StoreApiWithPersist } from 'zustand/middleware'

export const ONBOARDING_TOUR_LOCAL_STORAGE_KEY = 'getting-started-tour'

export enum GettingStartedTourLanguage {
    C = 'C',
    Go = 'Go',
    Java = 'Java',
    Javascript = 'JavaScript',
    Php = 'PHP',
    Python = 'Python',
    Typescript = 'TypeScript',
}

export interface GettingStartedTourState {
    completedIDs?: string[]
    continueID?: string
    status?: 'steps' | 'languages' | 'closed' | 'completed'
    language?: GettingStartedTourLanguage

    addCompletedID: (id: string) => void
    setLanguage: (id: GettingStartedTourState['language']) => void
    setLanguageStatus: (continueID: string) => void
    restart: () => void
    close: () => void
    complete: () => void
}

export const useGettingStartedTourState = create<GettingStartedTourState>(
    persist<
        GettingStartedTourState,
        SetState<GettingStartedTourState>,
        GetState<GettingStartedTourState>,
        StoreApiWithPersist<GettingStartedTourState>
    >(
        (set, get): GettingStartedTourState => ({
            status: 'steps',
            setLanguage: language => set({ language, status: 'steps' }),
            setLanguageStatus: continueID => set({ status: 'languages', continueID }),
            addCompletedID: id => set({ completedIDs: [...(get().completedIDs ?? []), id] }),
            close: () => set({ status: 'closed' }),
            complete: () => set({ status: 'completed' }),
            restart: () => set({ completedIDs: [], status: 'steps', continueID: undefined, language: undefined }),
        }),
        {
            name: ONBOARDING_TOUR_LOCAL_STORAGE_KEY,
        }
    )
)
