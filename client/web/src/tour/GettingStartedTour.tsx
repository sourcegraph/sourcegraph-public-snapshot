import { type FC, memo } from 'react'

import { withFeatureFlag } from '../featureFlags/withFeatureFlag'

import { Tour, type TourProps } from './components/Tour/Tour'
import { TourInfo } from './components/Tour/TourInfo'
import { withErrorBoundary } from './components/withErrorBoundary'
import { authenticatedExtraTask, useOnboardingTasks } from './data'

const GatedTour = withFeatureFlag('end-user-onboarding', Tour)

type TourWrapperProps = Omit<TourProps, 'useStore' | 'eventPrefix' | 'tasks' | 'id' | 'defaultSnippets'>

const TourWrapper: FC<TourWrapperProps> = props => {
    const { loading, error, data } = useOnboardingTasks()
    if (loading || error || !data) {
        return null
    }

    return (
        <GatedTour
            {...props}
            id="TourAuthenticated"
            tasks={data.tasks}
            defaultSnippets={data.defaultSnippets}
            extraTask={authenticatedExtraTask}
        />
    )
}

// This needed to be split up into two compontent definitions because
// eslint warns that `useOnboardingTasks` cannot be used inside a callback
// (but the value passed to `withErrorBoundary` really is a component)
const TourWithErrorBoundary = memo(withErrorBoundary(TourWrapper))

export const GettingStartedTour = Object.assign(TourWithErrorBoundary, {
    Info: TourInfo,
})
