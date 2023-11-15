import { Navigate, useLocation } from 'react-router-dom'

import { TelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { userURL } from '..'
import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'

/**
 * Redirects from /settings to /user/$USERNAME/settings, where $USERNAME is the currently authenticated user's
 * username.
 */
export const RedirectToUserSettings = withAuthenticatedUser<{
    authenticatedUser: AuthenticatedUser
    telemetryRecorder: TelemetryRecorder
}>(({ authenticatedUser }) => {
    const location = useLocation()

    return (
        <Navigate
            to={{
                pathname: `${userURL(authenticatedUser.username)}/settings`,
                search: location.search,
                hash: location.hash,
            }}
            replace={true}
        />
    )
})
