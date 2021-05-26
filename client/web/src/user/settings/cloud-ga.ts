import { AuthenticatedUser } from '../../auth'
import { UserAreaUserFields } from '../../graphql-operations'

export interface UserProps {
    user: Pick<UserAreaUserFields, 'id' | 'tags' | 'builtinAuth'>
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'tags'>
}

export const externalServiceUserMode = (props: UserProps): 'disabled' | 'public' | 'all' | 'unknown' =>
{
    const siteMode = window.context.externalServicesUserMode
    if (siteMode === 'all') {
        // Site mode already allows all repo types, no need to check user tags
        return siteMode
    }
    if (props.user.tags?.some(tag => tag === 'AllowUserExternalServicePrivate')) {
        return 'all'
    }
    if (props.user.tags?.some(tag => tag === 'AllowUserExternalServicePublic')) {
        return 'public'
    }
    return siteMode
}

export const userExternalServicesEnabled = (props: UserProps): boolean => {
    const mode = externalServiceUserMode(props)
    return mode === 'all' || mode === 'public'
}

export const showPasswordsPage = (props: UserProps): boolean =>
{
    // user is signed-in with builtin Auth and External Service is not public
    const mode = externalServiceUserMode(props)
    return props.user.builtinAuth && (mode === 'disabled' || mode === 'unknown')
}

export const showAccountSecurityPage = (props: UserProps): boolean => !showPasswordsPage(props)
