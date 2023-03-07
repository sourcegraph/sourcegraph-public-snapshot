import * as React from 'react'

import { Routes, Route } from 'react-router-dom'

import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { ErrorAlert, LoadingSpinner } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { RouteError } from '../components/ErrorBoundary'
import { NotFoundPage } from '../components/HeroPage'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

import type { TeamAreaProps } from './area/TeamArea'
import type { TeamListPageProps } from './list/TeamListPage'
import type { NewTeamPageProps } from './new/NewTeamPage'

const TeamArea = lazyComponent<TeamAreaProps, 'TeamArea'>(() => import('./area/TeamArea'), 'TeamArea')
const TeamListPage = lazyComponent<TeamListPageProps, 'TeamListPage'>(
    () => import('./list/TeamListPage'),
    'TeamListPage'
)
const NewTeamPage = lazyComponent<NewTeamPageProps, 'NewTeamPage'>(() => import('./new/NewTeamPage'), 'NewTeamPage')

export interface Props {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display team-related pages.
 */
const AuthenticatedTeamsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [enableTeams, fetchStatus, fetchError] = useFeatureFlag('search-ownership')

    switch (fetchStatus) {
        case 'initial':
            return <LoadingSpinner className="m-2" />
        case 'error':
            return <ErrorAlert prefix="Failed to load teams feature flag" error={fetchError} />
    }

    // No teams on sourcegraph.com
    if (!enableTeams || props.isSourcegraphDotCom) {
        return <NotFoundPage pageType="team" />
    }
    return (
        <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
            <Routes>
                <Route path="new" element={<NewTeamPage />} errorElement={<RouteError />} />
                <Route path="" element={<TeamListPage {...props} />} errorElement={<RouteError />} />
                <Route path=":teamName/*" element={<TeamArea {...props} />} errorElement={<RouteError />} />
                <Route element={<NotFoundPage pageType="team" />} errorElement={<RouteError />} />
            </Routes>
        </React.Suspense>
    )
}

export const TeamsArea = withAuthenticatedUser(AuthenticatedTeamsArea)
