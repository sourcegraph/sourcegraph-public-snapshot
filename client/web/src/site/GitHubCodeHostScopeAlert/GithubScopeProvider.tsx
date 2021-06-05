import React, { FunctionComponent, useState, useEffect, useCallback } from 'react'

import { queryExternalServicesScope } from '../../components/externalServices/backend'
import { ExternalServiceKind } from '../../graphql-operations'

import { GitHubScopeContext } from './GithubScopeContext'

interface Props {
    children: React.ReactNode
    authenticatedUser: { id: string; tags: string[] } | null
}

export const GithubCodeHostScopeProvider: FunctionComponent<Props> = ({ children, authenticatedUser }) => {
    const [scopes, setScopes] = useState<string[] | null>(null)

    const checkGitHubServiceScope = useCallback(async (): Promise<void> => {
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
        checkGitHubServiceScope().catch(() => {})
    }, [checkGitHubServiceScope])

    const { Provider } = GitHubScopeContext

    return <Provider value={scopes}>{children}</Provider>
}
