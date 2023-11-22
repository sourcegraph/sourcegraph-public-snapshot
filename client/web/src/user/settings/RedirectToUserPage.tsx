import { Navigate, useLocation } from 'react-router-dom'

import { userURL } from '..'
import type { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'

/**
 * Redirects from /user/$PATH to /user/$USERNAME/$PATH, where $USERNAME is the currently
 * authenticated user's username.
 */
export const RedirectToUserPage = withAuthenticatedUser<{ authenticatedUser: AuthenticatedUser }>(
    ({ authenticatedUser }) => {
        const location = useLocation()
        const path = location.pathname.replace(/^\/user\/?/, '') // trim leading '/user/?'

        return (
            <Navigate
                replace={true}
                to={{
                    pathname: `${userURL(authenticatedUser.username)}/${path}`,
                    search: location.search,
                    hash: location.hash,
                }}
            />
        )
    }
)
