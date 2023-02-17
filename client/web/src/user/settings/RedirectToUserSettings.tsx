import { Navigate, useLocation } from 'react-router-dom-v5-compat'

import { userURL } from '..'
import { AuthenticatedUser } from '../../auth'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'

/**
 * Redirects from /settings to /user/$USERNAME/settings, where $USERNAME is the currently authenticated user's
 * username.
 */
export const RedirectToUserSettings = withAuthenticatedUser<{
    authenticatedUser: AuthenticatedUser
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
