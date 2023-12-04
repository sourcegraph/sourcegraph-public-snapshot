import create from 'zustand'

interface CodeInsightsLicenseState {
    licensed: boolean
    insightsLimit: number | null
}

const defaultState: CodeInsightsLicenseState = {
    licensed: false,
    insightsLimit: 2,
}

export const useCodeInsightsLicenseState = create<CodeInsightsLicenseState>(() => defaultState)
