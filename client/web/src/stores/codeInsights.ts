import create from 'zustand'

interface CodeInsightsState {
    licensed: boolean
    insightsLimit: number | null
}

const defaultState: CodeInsightsState = {
    licensed: false,
    insightsLimit: 2,
}

export const useCodeInsightsState = create<CodeInsightsState>(() => defaultState)
