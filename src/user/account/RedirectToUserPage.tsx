import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router'
import { userURL } from '..'
import * as GQL from '../../backend/graphqlschema'

/**
 * Redirects from /user/$PATH to /user/$USERNAME/$PATH, where $USERNAME is the currently
 * authenticated user's username.
 */
export const RedirectToUserPage: React.SFC<{
    user: GQL.IUser | null
    location: H.Location
}> = ({ user, location }) => {
    // If not logged in, redirect to sign in
    if (!user) {
        const newURL = new URL(window.location.href)
        newURL.pathname = '/sign-in'
        // Return to the current page after sign up/in.
        newURL.searchParams.set('returnTo', window.location.href)
        return <Redirect to={{ pathname: newURL.pathname, search: newURL.search }} />
    }

    const path = location.pathname.replace(/^\/user\//, '') // trim leading '/user/'
    return <Redirect to={{ pathname: `${userURL(user.username)}/${path}`, search: location.search }} />
}
