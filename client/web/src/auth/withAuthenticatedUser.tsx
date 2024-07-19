import React, { useEffect, type FunctionComponent, type PropsWithChildren } from 'react'

import { useLocation, useNavigate, type Location, type NavigateFunction } from 'react-router-dom'

import type { AuthenticatedUser } from '../auth'

interface WithAuthenticatedUserProps<U extends {} = AuthenticatedUser> {
    authenticatedUser: U
}

/**
 * Wraps a React component and requires an authenticated user. If the viewer is not authenticated, it redirects to
 * the sign-in flow.
 */
export function withAuthenticatedUser<P extends WithAuthenticatedUserProps<U>, U extends {} = AuthenticatedUser>(
    Component: React.ComponentType<P>
): React.FunctionComponent<
    Omit<P, 'authenticatedUser'> & {
        authenticatedUser: U | null
    }
> {
    // It's important to add names to all components to avoid full reload on hot-update.
    return function WithAuthenticatedUser({ authenticatedUser, ...props }) {
        const navigate = useNavigate()
        const location = useLocation()

        if (useRedirectToSignIn(authenticatedUser, navigate, location)) {
            return null
        }

        return <Component {...({ ...props, authenticatedUser } as P)} />
    }
}

export const AuthenticatedUserOnly: FunctionComponent<
    PropsWithChildren<{
        authenticatedUser: Pick<AuthenticatedUser, 'id'> | null
    }>
> = ({ authenticatedUser, children }) => {
    const navigate = useNavigate()
    const location = useLocation()

    if (useRedirectToSignIn(authenticatedUser, navigate, location)) {
        return null
    }

    return <>{children}</>
}

function useRedirectToSignIn<U extends {} = AuthenticatedUser>(
    authenticatedUser: U | null,
    navigate: NavigateFunction,
    location: Location
): boolean {
    useEffect(() => {
        // If not logged in, redirect to sign in.
        if (!authenticatedUser) {
            // Return to the current page after sign up/in.
            const returnTo = `${location.pathname}${location.search}${location.hash}`
            navigate(`/sign-in?returnTo=${encodeURIComponent(returnTo)}`, { replace: true })
        }
    }, [authenticatedUser, navigate, location])

    return !authenticatedUser
}
