import React from 'react'

import { FeatureFlagProps } from '../featureFlags/featureFlags'

import { Tour, TourProps } from './components/Tour/Tour'
import { TourInfo } from './components/Tour/TourInfo'
import { withErrorBoundary } from './components/withErrorBoundary'
import { authenticatedExtraTask, authenticatedTasks, visitorsTasks } from './data'

type TourWithErrorBoundaryProps = Omit<TourProps, 'useStore' | 'eventPrefix' | 'tasks' | 'id'> &
    FeatureFlagProps & {
        isAuthenticated?: boolean
        isSourcegraphDotCom: boolean
    }

const TourWithErrorBoundary = withErrorBoundary(
    ({ isAuthenticated, featureFlags, isSourcegraphDotCom, ...props }: TourWithErrorBoundaryProps) => {
        // Do not show if on prem
        if (!isSourcegraphDotCom) {
            return null
        }

        // Show visitors version
        if (!isAuthenticated) {
            return <Tour {...props} id="Tour" tasks={visitorsTasks} />
        }

        // Show for enabled control group
        if (featureFlags.get('quick-start-tour-for-authenticated-users')) {
            return (
                <Tour {...props} id="TourAuthenticated" tasks={authenticatedTasks} extraTask={authenticatedExtraTask} />
            )
        }

        // Do not show for the rest
        return null
    }
)

export const GettingStartedTour = Object.assign(TourWithErrorBoundary, {
    Info: TourInfo,
})
