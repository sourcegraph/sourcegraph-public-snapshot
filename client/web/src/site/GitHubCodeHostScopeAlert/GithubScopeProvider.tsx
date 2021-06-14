import React, { FunctionComponent, useState, useEffect, useCallback, createContext, useContext } from 'react'

import { queryExternalServicesScope } from '../../components/externalServices/backend'
import { ExternalServiceKind } from '../../graphql-operations'

type Scopes = string[] | null
type SetScopes = (scopes: Scopes) => void

interface GitHubScopeContext {
    scopes: Scopes
    setScopes: SetScopes
}

const GitHubScopeContext = createContext<GitHubScopeContext | undefined>(undefined)

interface Props {
    children: React.ReactNode
    authenticatedUser: { id: string; tags: string[] } | null
}

export const GitHubServiceScopeProvider: FunctionComponent<Props> = ({ children, authenticatedUser }) => {
    const [scopes, setScopes] = useState<string[] | null>(null)

    const fetchGitHubServiceScope = useCallback(async (): Promise<void> => {
        if (authenticatedUser) {
            // fetch all code hosts for given user
            const { nodes: fetchedServices } = await queryExternalServicesScope({
                namespace: authenticatedUser.id,
            }).toPromise()

            // check if user has a GitHub code host
            const gitHubService = fetchedServices.find(({ kind }) => kind === ExternalServiceKind.GITHUB)

            if (gitHubService) {
                setScopes(gitHubService.grantedScopes)
            }
        }
    }, [authenticatedUser])

    useEffect(() => {
        fetchGitHubServiceScope().catch(() => {
            // there's no actionable information we can display here
        })
    }, [fetchGitHubServiceScope])

    const { Provider } = GitHubScopeContext

    return <Provider value={{ scopes, setScopes }}>{children}</Provider>
}

export const useGitHubScopeContext = (): GitHubScopeContext => {
    const context = useContext(GitHubScopeContext)

    if (context === undefined) {
        throw new Error('useCount must be used within a GitHubServiceScopeProvider')
    }

    return context
}
