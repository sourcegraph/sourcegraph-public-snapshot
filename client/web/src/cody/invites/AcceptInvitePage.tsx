import React, { useEffect } from 'react'

import { Navigate, useLocation } from 'react-router-dom'

import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'

import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import { CodyProRoutes } from '../codyProRoutes'

interface CodyAcceptInvitePageProps extends TelemetryV2Props {
    authenticatedUser: AuthenticatedUser
}

const AuthenticatedCodyAcceptInvitePage: React.FunctionComponent<CodyAcceptInvitePageProps> = ({
    telemetryRecorder,
}) => {
    const location = useLocation()

    useEffect(() => {
        telemetryRecorder.recordEvent('cody.invites.accept', 'view')
    }, [telemetryRecorder])

    // navigate to the manage page and passthrough the search params
    return <Navigate to={CodyProRoutes.Manage + location.search} replace={true} />
}

export const CodyAcceptInvitePage = withAuthenticatedUser(AuthenticatedCodyAcceptInvitePage)
