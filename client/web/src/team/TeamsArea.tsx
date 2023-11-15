import * as React from 'react'

import { Routes, Route } from 'react-router-dom'

import { TelemetryV2Props, noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../auth'
import { withAuthenticatedUser } from '../auth/withAuthenticatedUser'
import { RouteError } from '../components/ErrorBoundary'
import { NotFoundPage } from '../components/HeroPage'

import type { TeamAreaProps } from './area/TeamArea'
import type { TeamListPageProps } from './list/TeamListPage'
import type { NewTeamPageProps } from './new/NewTeamPage'

const TeamArea = lazyComponent<TeamAreaProps, 'TeamArea'>(() => import('./area/TeamArea'), 'TeamArea')
const TeamListPage = lazyComponent<TeamListPageProps, 'TeamListPage'>(
    () => import('./list/TeamListPage'),
    'TeamListPage'
)
const NewTeamPage = lazyComponent<NewTeamPageProps, 'NewTeamPage'>(() => import('./new/NewTeamPage'), 'NewTeamPage')

export interface Props extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
    isSourcegraphDotCom: boolean
}

/**
 * Renders a layout of a sidebar and a content area to display team-related pages.
 */
const AuthenticatedTeamsArea: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    // No teams on sourcegraph.com
    if (props.isSourcegraphDotCom) {
        return <NotFoundPage pageType="team" />
    }
    return (
        <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
            <Routes>
                <Route
                    path="new"
                    element={<NewTeamPage telemetryRecorder={noOpTelemetryRecorder} />}
                    errorElement={<RouteError />}
                />
                <Route path="" element={<TeamListPage {...props} />} errorElement={<RouteError />} />
                <Route path=":teamName/*" element={<TeamArea {...props} />} errorElement={<RouteError />} />
                <Route path="*" element={<NotFoundPage pageType="team" />} errorElement={<RouteError />} />
            </Routes>
        </React.Suspense>
    )
}

export const TeamsArea = withAuthenticatedUser(AuthenticatedTeamsArea)
