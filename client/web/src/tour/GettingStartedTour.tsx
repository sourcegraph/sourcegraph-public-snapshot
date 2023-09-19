import { memo } from 'react'

import { withFeatureFlag } from '../featureFlags/withFeatureFlag'

import { Tour, type TourProps } from './components/Tour/Tour'
import { TourInfo } from './components/Tour/TourInfo'
import { withErrorBoundary } from './components/withErrorBoundary'
import { authenticatedExtraTask, authenticatedTasks } from './data'

const GatedTour = withFeatureFlag('end-user-onboarding', Tour)

type TourWithErrorBoundaryProps = Omit<TourProps, 'useStore' | 'eventPrefix' | 'tasks' | 'id'>

const TourWithErrorBoundary = memo(
    withErrorBoundary(({ ...props }: TourWithErrorBoundaryProps) => (
        <GatedTour {...props} id="TourAuthenticated" tasks={authenticatedTasks} extraTask={authenticatedExtraTask} />
    ))
)

export const GettingStartedTour = Object.assign(TourWithErrorBoundary, {
    Info: TourInfo,
})
