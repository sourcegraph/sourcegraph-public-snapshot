import React from 'react'

import type { TourTaskStepType } from '@sourcegraph/shared/src/settings/temporary'

interface TourContextType {
    /**
     * Marks step as completed and triggers `${TourId}${step.id}Clicked` event log
     */
    onStepClick: (step: TourTaskStepType) => void

    /**
     * Restarts tour and triggers `${TourId}${step.id}Clicked` event log
     */
    onRestart: (step: TourTaskStepType) => void
}
export const TourContext = React.createContext<TourContextType>({} as TourContextType)
