import { UserAreaUserFields } from '../../graphql-operations'
import { AuthenticatedUser } from '../../auth'

export interface UserProps {
    user: Pick<UserAreaUserFields, 'id' | 'tags' | 'builtinAuth'>
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'tags'>
}

export const allowUserExternalServicePublic = (props: UserProps): boolean =>
    window.context.externalServicesUserModeEnabled ||
    (props.user.id === props.authenticatedUser.id &&
        props.authenticatedUser.tags.includes('AllowUserExternalServicePublic')) ||
    props.user.tags?.includes('AllowUserExternalServicePublic')

export const showPasswordsPage = (props: UserProps): boolean =>
    // user is signed-in with builtin Auth and External Service is not public
    props.user.builtinAuth && !allowUserExternalServicePublic(props)

export const showAccountSecurityPage = (props: UserProps): boolean => !showPasswordsPage(props)
