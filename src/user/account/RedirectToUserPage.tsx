import * as H from 'history'
import * as React from 'react'
import { Redirect } from 'react-router'
import { userURL } from '..'
import { withAuthenticatedUser } from '../../auth/withAuthenticatedUser'
import * as GQL from '../../backend/graphqlschema'

/**
 * Redirects from /user/$PATH to /user/$USERNAME/$PATH, where $USERNAME is the currently
 * authenticated user's username.
 */
export const RedirectToUserPage = withAuthenticatedUser(
    ({ authenticatedUser, location }: { authenticatedUser: GQL.IUser; location: H.Location }) => {
        const path = location.pathname.replace(/^\/user\//, '') // trim leading '/user/'
        return <Redirect to={{ pathname: `${userURL(authenticatedUser.username)}/${path}`, search: location.search }} />
    }
)
