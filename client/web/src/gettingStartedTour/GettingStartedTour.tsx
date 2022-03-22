import React from 'react'

import { FeatureFlagProps } from '../featureFlags/featureFlags'

import {
    authenticatedExtraTask,
    authenticatedTasks,
    authenticatedTasksAllUseCases,
    visitorsTasks,
} from './components/data'
import { Tour, TourProps } from './components/Tour'
import { withErrorBoundary } from './components/withErrorBoundary'

type GettingStartedTourProps = Omit<TourProps, 'useStore' | 'eventPrefix' | 'tasks' | 'id'> &
    FeatureFlagProps & {
        isAuthenticated?: boolean
        isSourcegraphDotCom: boolean
    }

// TODO: remove old components

export const GettingStartedTour = withErrorBoundary(
    ({ isAuthenticated, featureFlags, isSourcegraphDotCom, ...props }: GettingStartedTourProps) => {
        if (!isSourcegraphDotCom) {
            return null
        }

        if (!isAuthenticated) {
            return <Tour {...props} id="Tour" tasks={visitorsTasks} />
        }

        return (
            <Tour
                {...props}
                id="TourAuthenticated"
                tasks={featureFlags.get('quick-start-tour') ? authenticatedTasks : authenticatedTasksAllUseCases}
                extraTask={authenticatedExtraTask}
            />
        )
    }
)
