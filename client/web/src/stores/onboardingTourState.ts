import create, { GetState, SetState } from 'zustand'
import { persist, StoreApiWithPersist } from 'zustand/middleware'

export const ONBOARDING_TOUR_LOCAL_STORAGE_KEY = 'getting-started-tour'

export enum OnboardingTourLanguage {
    C = 'C',
    Go = 'Go',
    Java = 'Java',
    Javascript = 'JavaScript',
    Php = 'PHP',
    Python = 'Python',
    Typescript = 'TypeScript',
}

export interface OnboardingTourState {
    completedIDs?: string[]
    continueID?: string
    status?: 'steps' | 'languages' | 'closed' | 'completed'
    language?: OnboardingTourLanguage

    addCompletedID: (id: string) => void
    setLanguage: (id: OnboardingTourState['language']) => void
    setLanguageStatus: (continueID: string) => void
    restart: () => void
    close: () => void
    complete: () => void
}

export const useOnboardingTourState = create<OnboardingTourState>(
    persist<
        OnboardingTourState,
        SetState<OnboardingTourState>,
        GetState<OnboardingTourState>,
        StoreApiWithPersist<OnboardingTourState>
    >(
        (set, get): OnboardingTourState => ({
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
