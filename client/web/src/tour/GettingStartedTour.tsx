import { FC, memo, PropsWithChildren } from 'react'

import { withFeatureFlag } from '../featureFlags/withFeatureFlag'

import { Tour, type TourProps } from './components/Tour/Tour'
import { TourInfo } from './components/Tour/TourInfo'
import { withErrorBoundary } from './components/withErrorBoundary'
import {
    authenticatedExtraTask,
    authenticatedTasks,
    visitorsTasks,
    visitorsTasksWithNotebook,
    visitorsTasksWithNotebookExtraTask,
} from './data'
import { gql, useQuery } from "@sourcegraph/http-client";
import { OnboardingTourConfigResult, OnboardingTourConfigVariables } from "../graphql-operations";
import { LoadingSpinner } from "@sourcegraph/wildcard";

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

const ONBOARDING_TOUR_QUERY = gql`
    query OnboardingTourConfig {
        onboardingTourContent {
            current {
                id
                value
            }
        }
    }
`

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
        // Show visitors version
        if (!isAuthenticated) {
            return <TourVisitor {...props} />
        }



        return (
            <TourAuthenticated
                {...props}
                id="TourAuthenticated"
                extraTask={authenticatedExtraTask}
            />
        )
    })
)

interface Props {

}

export const TourAuth: FC<PropsWithChildren<Props>> = (props) => {
    const { data, loading, error, previousData } = useQuery<OnboardingTourConfigResult, OnboardingTourConfigVariables>(
        ONBOARDING_TOUR_QUERY,
        {}
    )

    return (
        <div>
            {!loading && <Tour
                {...props}
                tasks={JSON.parse(data?.onboardingTourContent.current?.value).tasks}
            />}

            {loading && <LoadingSpinner></LoadingSpinner>}
        </div>
    )
}



// Show for enabled control group
export const TourAuthenticated = withFeatureFlag('quick-start-tour-for-authenticated-users', TourAuth)

export const GettingStartedTour = Object.assign(TourWithErrorBoundary, {
    Info: TourInfo,
})
