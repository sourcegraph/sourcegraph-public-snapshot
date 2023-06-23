import { useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import { isEqual } from 'lodash'

import { ErrorLike } from '@sourcegraph/common'
import { useLazyQuery, useMutation, useQuery } from '@sourcegraph/http-client'

import {
    AddLocalRepositoriesResult,
    AddLocalRepositoriesVariables,
    AddRemoteCodeHostResult,
    AddRemoteCodeHostVariables,
    DeleteRemoteCodeHostResult,
    DeleteRemoteCodeHostVariables,
    DiscoverLocalRepositoriesResult,
    DiscoverLocalRepositoriesVariables,
    ExternalServiceKind,
    GetLocalCodeHostsResult,
    GetLocalCodeHostsVariables,
    LocalRepository,
} from '../../../graphql-operations'
import { ADD_CODE_HOST, ADD_LOCAL_REPOSITORIES, DELETE_CODE_HOST } from '../../queries'

import { LocalCodeHost, createDefaultLocalServiceConfig, getLocalServicePaths, getLocalServices } from './helpers'
import { DISCOVER_LOCAL_REPOSITORIES, GET_LOCAL_CODE_HOSTS } from './queries'

type Path = string

interface useNewLocalRepositoriesPathsAPI {
    loading: boolean
    loaded: boolean
    error: ErrorLike | undefined
    paths: Path[]
    addNewPaths: (paths: Path[]) => Promise<void>
    deletePath: (path: Path) => Promise<void>
}

export function useNewLocalRepositoriesPaths(): useNewLocalRepositoriesPathsAPI {
    const { data, previousData, loading, error } = useQuery<GetLocalCodeHostsResult>(GET_LOCAL_CODE_HOSTS, {
        fetchPolicy: 'cache-and-network',
    })

    const [getLocalRepositories] = useLazyQuery<DiscoverLocalRepositoriesResult, DiscoverLocalRepositoriesVariables>(
        DISCOVER_LOCAL_REPOSITORIES,
        { fetchPolicy: 'network-only' }
    )

    const apolloClient = useApolloClient()
    const [addLocalCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)
    const [deleteLocalCodeHost] = useMutation<DeleteRemoteCodeHostResult, DeleteRemoteCodeHostVariables>(
        DELETE_CODE_HOST
    )

    const addNewPaths = async (paths: Path[]): Promise<void> => {
        const { data } = await getLocalRepositories({ variables: { paths } })
        const repositoriesCount = data?.localDirectories.repositories.length ?? Infinity

        if (repositoriesCount > 1) {
            // Throw an error about multiple repositories
            return
        }

        for (const path of paths) {
            // Create a new local external service for this path
            await addLocalCodeHost({
                variables: {
                    input: {
                        displayName: `Local repositories service (${path})`,
                        config: createDefaultLocalServiceConfig(path),
                        kind: ExternalServiceKind.OTHER,
                    },
                },
            })
        }

        await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts'] })
    }

    const deletePath = async (path: Path): Promise<void> => {
        const localServices = getLocalServices(data, false)
        const localServiceToDelete = localServices.find(localService => localService.path === path)

        if (!localServiceToDelete) {
            return
        }

        await deleteLocalCodeHost({ variables: { id: localServiceToDelete.id } })
        await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts'] })
    }

    return {
        error,
        loading,
        addNewPaths,
        deletePath,
        loaded: !!data || !!previousData,
        paths: getLocalServicePaths(data),
    }
}

interface LocalRepositoriesPathAPI {
    loading: boolean
    loaded: boolean
    error: ErrorLike | undefined
    paths: Path[]
    autogeneratedPaths: Path[]
    setPaths: (newPaths: Path[]) => void
}

/**
 * Returns a list of local paths that we use to gather local repositories
 * from the user's machine. Internally, it stores paths with special type of
 * external service, if service doesn't exist it returns empty list.
 */
export function useLocalRepositoriesPaths(): LocalRepositoriesPathAPI {
    const apolloClient = useApolloClient()

    const [error, setError] = useState<ErrorLike | undefined>()
    const [paths, setPaths] = useState<string[]>([])

    const [addLocalCodeHost] = useMutation<AddRemoteCodeHostResult, AddRemoteCodeHostVariables>(ADD_CODE_HOST)

    const [deleteLocalCodeHost] = useMutation<DeleteRemoteCodeHostResult, DeleteRemoteCodeHostVariables>(
        DELETE_CODE_HOST
    )

    const { data, previousData, loading } = useQuery<GetLocalCodeHostsResult>(GET_LOCAL_CODE_HOSTS, {
        fetchPolicy: 'network-only',
        // Sync local external service paths on first load
        onCompleted: data => {
            setPaths(getLocalServicePaths(data))
        },
        onError: setError,
    })

    // Automatically creates or deletes local external service to
    // match user chosen paths for local repositories.
    useEffect(() => {
        if (loading) {
            return
        }

        const localServices = getLocalServices(data)
        const localServicePaths = getLocalServicePaths(data)
        const havePathsChanged = !isEqual(paths, localServicePaths)

        // Do nothing if paths haven't changed
        if (!havePathsChanged) {
            return
        }

        setError(undefined)

        async function syncExternalServices(): Promise<void> {
            // Create/update local external services
            for (const path of paths) {
                // If we already have a local external service for this path, skip it
                if (localServicePaths.includes(path)) {
                    continue
                }

                // Create a new local external service for this path
                await addLocalCodeHost({
                    variables: {
                        input: {
                            displayName: `Local repositories service (${path})`,
                            config: createDefaultLocalServiceConfig(path),
                            kind: ExternalServiceKind.OTHER,
                        },
                    },
                })
            }

            // Delete local external services that are no longer in the list
            for (const localService of localServices || []) {
                // If we still have a local external service for this path, skip it
                if (paths.includes(localService.path)) {
                    continue
                }

                // Delete local external service for this path
                await deleteLocalCodeHost({
                    variables: {
                        id: localService.id,
                    },
                })
            }

            // Refetch local external services and status after all mutations have been completed.
            await apolloClient.refetchQueries({ include: ['GetLocalCodeHosts', 'StatusAndRepoStats'] })
        }

        syncExternalServices().catch(setError)
    }, [data, paths, loading, apolloClient, addLocalCodeHost, deleteLocalCodeHost])

    return {
        error,
        loading,
        loaded: !!data || !!previousData,
        paths,
        autogeneratedPaths: getLocalServices(data, true).map(item => item.path),
        setPaths,
    }
}

interface LocalRepositoriesInput {
    paths: Path[]
    skip: boolean
}

interface LocalRepositoriesResult {
    loading: boolean
    error: ErrorLike | undefined
    loaded: boolean
    repositories: LocalRepository[]
}

const EMPTY_REPOSITORY_LIST: LocalRepository[] = []

/** Returns list of local repositories by a given list of local paths. */
export function useLocalRepositories({ paths, skip }: LocalRepositoriesInput): LocalRepositoriesResult {
    const { data, previousData, loading, error } = useQuery<
        DiscoverLocalRepositoriesResult,
        DiscoverLocalRepositoriesVariables
    >(DISCOVER_LOCAL_REPOSITORIES, {
        skip,
        variables: { paths },
        fetchPolicy: 'network-only',
    })

    return {
        loading,
        error,
        loaded: skip || !!data || !!previousData,
        repositories:
            data?.localDirectories?.repositories ??
            previousData?.localDirectories?.repositories ??
            EMPTY_REPOSITORY_LIST,
    }
}

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
