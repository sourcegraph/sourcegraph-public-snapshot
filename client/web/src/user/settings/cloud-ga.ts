import { AuthenticatedUser } from '../../auth'
import { UserAreaUserFields, Scalars } from '../../graphql-operations'

type Scopes = string[] | null

export interface UserProps {
    user: Pick<UserAreaUserFields, 'id' | 'tags' | 'builtinAuth'>
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'tags'>
}

export interface Owner {
    id: Scalars['ID']
    type: 'user' | 'org'
    tags?: string[]
    name?: string
}

export const externalServiceUserMode = (props: UserProps): 'disabled' | 'public' | 'all' | 'unknown' =>
    externalServiceUserModeFromTags(props.user.tags || [])

export const showPasswordsPage = (props: UserProps): boolean => {
    // user is signed-in with builtin Auth and External Service is not public
    const mode = externalServiceUserMode(props)
    return props.user.builtinAuth && (mode === 'disabled' || mode === 'unknown')
}

export const showAccountSecurityPage = (props: UserProps): boolean => !showPasswordsPage(props)

export const externalServiceUserModeFromTags = (tags?: string[]): 'disabled' | 'public' | 'all' | 'unknown' => {
    const siteMode = window.context.externalServicesUserMode
    if (siteMode === 'all') {
        // Site mode already allows all repo types, no need to check user tags
        return siteMode
    }
    if (tags?.includes('AllowUserExternalServicePrivate')) {
        return 'all'
    }
    if (tags?.includes('AllowUserExternalServicePublic')) {
        return 'public'
    }
    return siteMode
}

// If the user is allowed to add private code but they don't have the 'repo' scope
// then we need to request it.
export const githubRepoScopeRequired = (tags: string[] = [], scopes?: Scopes): boolean =>
    requiredScope('repo', tags, scopes)

// If the user is allowed to add private code but they don't have the 'api' scope
// then we need to request it.
export const gitlabAPIScopeRequired = (tags: string[] = [], scopes?: Scopes): boolean =>
    requiredScope('api', tags, scopes)

const requiredScope = (scope: string, tags: string[], scopes?: Scopes): boolean => {
    const allowedPrivate = externalServiceUserModeFromTags(tags) === 'all'
    if (!Array.isArray(scopes)) {
        return false
    }
    return allowedPrivate && !scopes.includes(scope)
}
