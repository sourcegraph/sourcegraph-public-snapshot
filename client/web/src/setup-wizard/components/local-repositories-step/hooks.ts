import { useCallback, useMemo } from 'react'

import { useApolloClient } from '@apollo/client'

import type { ErrorLike } from '@sourcegraph/common'
import { useMutation, useQuery } from '@sourcegraph/http-client'

import type {
    AddLocalRepositoriesResult,
    AddLocalRepositoriesVariables,
    DeleteRemoteCodeHostResult,
    DeleteRemoteCodeHostVariables,
    GetLocalCodeHostsResult,
    GetLocalCodeHostsVariables,
} from '../../../graphql-operations'
import { ADD_LOCAL_REPOSITORIES, DELETE_CODE_HOST } from '../../queries'

import type { LocalCodeHost } from './helpers'
import { GET_LOCAL_CODE_HOSTS } from './queries'

interface LocalCodeHostResult {
    loading: boolean
    error: ErrorLike | undefined
    loaded: boolean
    services: LocalCodeHost[]
    addRepositories: (paths: string[]) => Promise<void>
    deleteService: (service: LocalCodeHost) => Promise<void>
}

const EMPTY_CODEHOST_LIST: LocalCodeHost[] = []

export function useLocalExternalServices(): LocalCodeHostResult {
    const apolloClient = useApolloClient()

    const { data, previousData, loading, error } = useQuery<GetLocalCodeHostsResult, GetLocalCodeHostsVariables>(
        GET_LOCAL_CODE_HOSTS,
        {
            fetchPolicy: 'network-only',
        }
    )

    const [addLocalRepositories] = useMutation<AddLocalRepositoriesResult, AddLocalRepositoriesVariables>(
        ADD_LOCAL_REPOSITORIES
    )

    const [deleteLocalCodeHost] = useMutation<DeleteRemoteCodeHostResult, DeleteRemoteCodeHostVariables>(
        DELETE_CODE_HOST
    )

    const addRepositories = useCallback(
        async (paths: string[]): Promise<void> => {
            await addLocalRepositories({ variables: { paths } })
            await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts'] })
        },
        [addLocalRepositories, apolloClient]
    )

    const deleteService = useCallback(
        async (service: LocalCodeHost): Promise<void> => {
            await deleteLocalCodeHost({ variables: { id: service.id } })
            await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts'] })
        },
        [deleteLocalCodeHost, apolloClient]
    )

    const services = data?.localExternalServices ?? previousData?.localExternalServices ?? EMPTY_CODEHOST_LIST

    return {
        loading,
        error,
        loaded: !!data || !!previousData,
        // TODO: Determine folder/single repo on the server. Just comparing the length is not technically correct.
        // (a folder might have only one repository)
        services: useMemo(
            () => services.map(service => ({ ...service, isFolder: service.repositories.length !== 1 })),
            [services]
        ),
        deleteService,
        addRepositories,
    }
}
