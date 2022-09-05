import React from 'react'

import { TourLanguage, TourTaskStepType } from './types'

interface TourContextType {
    /**
     * Language user has chosen initially for the tour
     */
    language?: TourLanguage

    /**
     * Stores context.language and triggers `${TourId}LanguageClicked` event log
     */
    onLanguageSelect: (language: TourLanguage) => void

    /**
     * Marks step as completed and triggers `${TourId}${step.id}Clicked` event log
     */
    onStepClick: (step: TourTaskStepType, language?: TourLanguage) => void

    /**
     * Restarts tour and triggers `${TourId}${step.id}Clicked` event log
     */
    onRestart: (step: TourTaskStepType) => void
}
export const TourContext = React.createContext<TourContextType>({} as TourContextType)
