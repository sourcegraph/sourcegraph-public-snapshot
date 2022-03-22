import React from 'react'

import { TourLanguage, TourTaskStepType } from './types'

interface TourContextType {
    language?: TourLanguage
    onLanguageSelect: (language: TourLanguage) => void
    onStepClick: (step: TourTaskStepType, language?: TourLanguage) => void
    onRestart: (step: TourTaskStepType) => void
}
export const TourContext = React.createContext<TourContextType>({})
