import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router'
import { userURL } from '..'
import * as GQL from '../../backend/graphqlschema'

/**
 * Redirects from /settings to /user/$USERNAME/settings, where $USERNAME is the currently authenticated user's
 * username.
 */
export const RedirectToUserSettings: React.SFC<{
    user: GQL.IUser | null
    location: H.Location
}> = ({ user, location }) => {
    // If not logged in, redirect to sign in
    if (!user) {
        const newURL = new URL(window.location.href)
        newURL.pathname = location.pathname === '/settings/accept-invite' ? '/sign-up' : '/sign-in'
        // Return to the current page after sign up/in.
        newURL.searchParams.set('returnTo', window.location.href)
        return <Redirect to={{ pathname: newURL.pathname, search: newURL.search }} />
    }

    // Append full location.pathname instead of just "/settings" so that we support /settings/accept-invite
    // redirect, too.
    return <Redirect to={{ pathname: `${userURL(user.username)}${location.pathname}`, search: location.search }} />
}
