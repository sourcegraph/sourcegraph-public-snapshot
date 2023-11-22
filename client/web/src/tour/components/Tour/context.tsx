import React from 'react'

import type { TourTaskStepType } from '@sourcegraph/shared/src/settings/temporary'
import type { UserOnboardingConfig } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'

interface TourContextType {
    /**
     * Marks step as completed and triggers `${TourId}${step.id}Clicked` event log
     */
    onStepClick: (step: TourTaskStepType) => void

    /**
     * Restarts tour and triggers `${TourId}${step.id}Clicked` event log
     */
    onRestart: (step: TourTaskStepType) => void

    /**
     * Provides user specific values for dynamic queries.
     */
    readonly userInfo?: UserOnboardingConfig['userinfo']

    /**
     * The function to use for determining whether queries return results.
     */
    isQuerySuccessful: (query: string) => Promise<boolean>
}
export const TourContext = React.createContext<TourContextType>({} as TourContextType)
