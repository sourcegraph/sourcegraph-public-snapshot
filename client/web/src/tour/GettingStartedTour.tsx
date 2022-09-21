import { memo } from 'react'

import { withFeatureFlag } from '../featureFlags/withFeatureFlag'

import { Tour, TourProps } from './components/Tour/Tour'
import { TourInfo } from './components/Tour/TourInfo'
import { withErrorBoundary } from './components/withErrorBoundary'
import {
    authenticatedExtraTask,
    authenticatedTasks,
    visitorsTasks,
    visitorsTasksWithNotebook,
    visitorsTasksWithNotebookExtraTask,
} from './data'

function TourVisitorWithNotebook(props: Omit<TourProps, 'tasks' | 'id'>): JSX.Element {
    return (
        <Tour
            {...props}
            id="TourWithNotebook"
            title="Code search basics"
            keepCompletedTasks={true}
            tasks={visitorsTasksWithNotebook}
            extraTask={visitorsTasksWithNotebookExtraTask}
        />
    )
}

function TourVisitorRegular(props: Omit<TourProps, 'tasks' | 'id'>): JSX.Element {
    return <Tour {...props} id="Tour" tasks={visitorsTasks} />
}

const TourVisitor = withFeatureFlag('ab-visitor-tour-with-notebooks', TourVisitorWithNotebook, TourVisitorRegular)

type TourWithErrorBoundaryProps = Omit<TourProps, 'useStore' | 'eventPrefix' | 'tasks' | 'id'> & {
    isAuthenticated?: boolean
    isSourcegraphDotCom: boolean
}

const TourWithErrorBoundary = memo(
    withErrorBoundary(({ isAuthenticated, isSourcegraphDotCom, ...props }: TourWithErrorBoundaryProps) => {
        // Do not show if on prem
        if (!isSourcegraphDotCom) {
            return null
        }

        // Show visitors version
        if (!isAuthenticated) {
            return <TourVisitor {...props} />
        }

        return (
            <TourAuthenticated
                {...props}
                id="TourAuthenticated"
                tasks={authenticatedTasks}
                extraTask={authenticatedExtraTask}
            />
        )
    })
)

// Show for enabled control group
export const TourAuthenticated = withFeatureFlag('quick-start-tour-for-authenticated-users', Tour)

export const GettingStartedTour = Object.assign(TourWithErrorBoundary, {
    Info: TourInfo,
})
