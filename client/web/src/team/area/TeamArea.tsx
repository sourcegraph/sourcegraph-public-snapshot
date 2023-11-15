import * as React from 'react'

import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Route, Routes, useParams } from 'react-router-dom'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { LoadingSpinner, ErrorMessage } from '@sourcegraph/wildcard'

import type { AuthenticatedUser } from '../../auth'
import { RouteError } from '../../components/ErrorBoundary'
import { HeroPage } from '../../components/HeroPage'
import type { TeamAreaTeamFields } from '../../graphql-operations'
import type { RouteV6Descriptor } from '../../util/contributions'

import { useTeam } from './backend'
import type { TeamChildTeamsPageProps } from './TeamChildTeamsPage'
import type { TeamMembersPageProps } from './TeamMembersPage'
import type { TeamProfilePageProps } from './TeamProfilePage'

const TeamProfilePage = lazyComponent<TeamProfilePageProps, 'TeamProfilePage'>(
    () => import('./TeamProfilePage'),
    'TeamProfilePage'
)
const TeamMembersPage = lazyComponent<TeamMembersPageProps, 'TeamMembersPage'>(
    () => import('./TeamMembersPage'),
    'TeamMembersPage'
)
const TeamChildTeamsPage = lazyComponent<TeamChildTeamsPageProps, 'TeamChildTeamsPage'>(
    () => import('./TeamChildTeamsPage'),
    'TeamChildTeamsPage'
)

const NotFoundPage: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <HeroPage icon={MapSearchIcon} title="404: Not Found" subtitle="Sorry, the requested team was not found." />
)

export interface TeamAreaRoute extends RouteV6Descriptor<TeamAreaRouteContext> {}

export interface TeamAreaProps {
    /**
     * The currently authenticated user.
     */
    authenticatedUser: AuthenticatedUser
}

/**
 * Properties passed to all page components in the team area.
 */
export interface TeamAreaRouteContext {
    /** The team that is the subject of the page. */
    team: TeamAreaTeamFields

    /** Called when the team is updated and must be reloaded. */
    onTeamUpdate: () => void

    /** The currently authenticated user. */
    authenticatedUser: AuthenticatedUser
}

export const TeamArea: React.FunctionComponent<TeamAreaProps> = ({ authenticatedUser }) => {
    const { teamName } = useParams<{ teamName: string }>()

    const { data, loading, error, refetch } = useTeam(teamName!)

    if (loading) {
        return null
    }
    if (error) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={error} />} />
    }

    if (!data?.team) {
        return (
            <HeroPage
                icon={AlertCircleIcon}
                title="Error"
                subtitle={<ErrorMessage error={new Error(`Team not found: ${JSON.stringify(teamName)}`)} />}
            />
        )
    }

    const context: TeamAreaRouteContext = {
        authenticatedUser,
        team: data.team,
        onTeamUpdate: refetch,
    }

    return (
        <React.Suspense fallback={<LoadingSpinner className="m-2" />}>
            <Routes>
                <Route
                    path=""
                    element={<TeamProfilePage {...context} telemetryRecorder={noOpTelemetryRecorder} />}
                    errorElement={<RouteError />}
                />
                <Route
                    path="members"
                    element={<TeamMembersPage {...context} telemetryRecorder={noOpTelemetryRecorder} />}
                    errorElement={<RouteError />}
                />
                <Route
                    path="child-teams"
                    element={<TeamChildTeamsPage {...context} telemetryRecorder={noOpTelemetryRecorder} />}
                    errorElement={<RouteError />}
                />
                <Route path="*" element={<NotFoundPage />} />
            </Routes>
        </React.Suspense>
    )
}
