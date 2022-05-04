import React, { FunctionComponent, useState, useEffect, useCallback, createContext, useContext } from 'react'

import { queryExternalServicesScope } from '../../components/externalServices/backend'
import { ExternalServiceKind } from '../../graphql-operations'

type Scope = string[] | null

interface Scopes {
    github?: Scope
    gitlab?: Scope
}
type SetScope = (kind: ExternalServiceKind.GITHUB | ExternalServiceKind.GITLAB, scope: Scope) => void

interface CodeHostScopeContext {
    scopes: Scopes
    setScope: SetScope
}

const CodeHostScopeContext = createContext<CodeHostScopeContext | undefined>(undefined)

interface Props {
    children: React.ReactNode
    authenticatedUser: { id: string; tags: string[] } | null
}

export const CodeHostScopeProvider: FunctionComponent<React.PropsWithChildren<Props>> = ({
    children,
    authenticatedUser,
}) => {
    const [scopes, setScopes] = useState<Scopes>({})

    const fetchCodeHostScope = useCallback(async (): Promise<void> => {
        if (authenticatedUser) {
            // fetch all code hosts for given user
            const { nodes: fetchedServices } = await queryExternalServicesScope({
                namespace: authenticatedUser.id,
            }).toPromise()

            // In theory users should have at most one of each
            const gitHubService = fetchedServices.find(({ kind }) => kind === ExternalServiceKind.GITHUB)
            const gitLabService = fetchedServices.find(({ kind }) => kind === ExternalServiceKind.GITLAB)

            const newScopes: Scopes = {}

            if (gitHubService) {
                newScopes.github = gitHubService.grantedScopes
            }
            if (gitLabService) {
                newScopes.gitlab = gitLabService.grantedScopes
            }

            if (gitHubService || gitLabService) {
                setScopes(newScopes)
            }
        }
    }, [authenticatedUser])

    const setScope: SetScope = (kind, scope): void => {
        setScopes(previousScope => ({
            ...previousScope,
            [kind.toLowerCase()]: scope,
        }))
    }

    useEffect(() => {
        fetchCodeHostScope().catch(() => {
            // there's no actionable information we can display here
        })
    }, [fetchCodeHostScope])

    const { Provider } = CodeHostScopeContext

    return <Provider value={{ scopes, setScope }}>{children}</Provider>
}

export const useCodeHostScopeContext = (): CodeHostScopeContext => {
    const context = useContext(CodeHostScopeContext)

    if (context === undefined) {
        throw new Error('CodeHostScopeContext must be used within a CodeHostScopeProvider')
    }

    return context
}
