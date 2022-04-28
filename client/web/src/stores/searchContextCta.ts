import create from 'zustand'

import { register } from './dispatcher'

import { useIsSourcegraphDotCom, useAuthenticatedUser } from '.'

interface SearchContextCtaState {
    hasUserAddedExternalServices: boolean
    hasUserAddedRepositories: boolean
    hasUserSyncedPublicRepositories: boolean
}

export const useSearchContextCta = create<SearchContextCtaState>(() => ({
    hasUserAddedExternalServices: false,
    hasUserAddedRepositories: false,
    hasUserSyncedPublicRepositories: false,
}))

register(event => {
    switch (event.type) {
        case 'UserExternalServicesOrRepositoriesUpdate':
            useSearchContextCta.setState(state => ({
                hasUserAddedRepositories:
                    event.userRepoCount !== undefined ? event.userRepoCount > 0 : state.hasUserAddedRepositories,
                hasUserAddedExternalServices: event.externalServicesCount > 0,
            }))
            break
        case 'SyncedPublicRepositoriesUpdate':
            useSearchContextCta.setState({
                hasUserSyncedPublicRepositories: event.count > 0,
            })
            break
    }
})

export function useShowSearchContextCta(): boolean {
    const isSourcegraphDotCom = useIsSourcegraphDotCom()
    const authenticatedUser = useAuthenticatedUser()
    const { hasUserAddedRepositories, hasUserSyncedPublicRepositories } = useSearchContextCta()
    const isUserAnOrgMember = authenticatedUser && authenticatedUser.organizations.nodes.length !== 0

    return isSourcegraphDotCom && !isUserAnOrgMember && !(hasUserAddedRepositories || hasUserSyncedPublicRepositories)
}
