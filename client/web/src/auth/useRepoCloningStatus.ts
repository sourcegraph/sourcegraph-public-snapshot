import { ApolloError, QueryLazyOptions } from '@apollo/client'
import { upperFirst } from 'lodash'

import { useLazyQuery, gql } from '@sourcegraph/shared/src/graphql/graphql'

import { Maybe, Exact, UserRepositoriesVariables } from '../graphql-operations'

interface UseRepoCloningStatusArguments {
    userId: string
    pollInterval: number
}

interface RepoLine {
    id: string
    title: string
    details: string
    progress: number
}

interface RepoCloningStatus {
    trigger: (options?: QueryLazyOptions<Exact<UserRepositoriesVariables>>) => void
    repos: RepoLine[] | undefined
    loading: boolean
    isDoneCloning: boolean
    error: ApolloError | undefined
}

interface RepoFields {
    id: string
    name: string
    mirrorInfo: { cloned: boolean; cloneProgress: string; cloneInProgress: boolean; updatedAt: Maybe<string> }
}

interface CloneProgressResult {
    node: Maybe<{
        repositories: {
            nodes: RepoFields[]
        }
    }>
}

// temp object to store previous cloning progress percentage
let previousPercentage: { [key: string]: number } = {}

const USER_AFFILIATED_REPOS_MIRROR_INFO = gql`
    query UserRepositoriesMirrorInfo(
        $id: ID!
        $first: Int
        $query: String
        $cloned: Boolean
        $notCloned: Boolean
        $indexed: Boolean
        $notIndexed: Boolean
        $externalServiceID: ID
    ) {
        node(id: $id) {
            ... on User {
                repositories(
                    first: $first
                    query: $query
                    cloned: $cloned
                    notCloned: $notCloned
                    indexed: $indexed
                    notIndexed: $notIndexed
                    externalServiceID: $externalServiceID
                ) {
                    nodes {
                        id
                        name
                        mirrorInfo {
                            cloned
                            cloneInProgress
                            cloneProgress
                            updatedAt
                        }
                    }
                }
            }
        }
    }
`

export const useRepoCloningStatus = ({
    userId,
    pollInterval = 5000,
}: UseRepoCloningStatusArguments): RepoCloningStatus => {
    let shouldStopPolling = true

    const [trigger, { called, data, loading, error, stopPolling }] = useLazyQuery<
        CloneProgressResult,
        UserRepositoriesVariables
    >(USER_AFFILIATED_REPOS_MIRROR_INFO, {
        variables: {
            id: userId,
            cloned: true,
            notCloned: true,
            indexed: true,
            notIndexed: true,
            first: null,
            query: null,
            externalServiceID: null,
        },
        pollInterval,
    })

    const repos = data?.node?.repositories.nodes

    if (!Array.isArray(repos)) {
        return {
            trigger,
            repos: undefined,
            isDoneCloning: false,
            loading,
            error,
        }
    }

    const repoLines = repos.reduce((lines, { id, name, mirrorInfo }) => {
        const { details, progress, cloned } = parseMirrorInfo(id, mirrorInfo)

        if (!cloned) {
            shouldStopPolling = false
        }

        lines.push({
            id,
            details,
            progress,
            title: ignoreExternalService(name),
        })

        return lines
    }, [] as RepoLine[])

    repoLines.sort((lineA, lineB) => lineB.progress - lineA.progress)

    // stop polling and cleanup memory when all repos are done cloning
    if (called && stopPolling && shouldStopPolling) {
        stopPolling()
        previousPercentage = {}
    }

    return { trigger, repos: repoLines, isDoneCloning: shouldStopPolling, loading, error }
}

const parseMirrorInfo = (
    id: string,
    mirrorInfo: RepoFields['mirrorInfo']
): { progress: number; details: string; cloned: boolean } => {
    const { cloneProgress, cloned, cloneInProgress } = mirrorInfo

    // cloned
    if (cloned) {
        return {
            cloned,
            details: 'Successfully cloned',
            progress: 100,
        }
    }

    // not cloned and cloning is not in progress
    if (!cloneInProgress) {
        return {
            cloned,
            details: 'Not cloned yet',
            progress: 0,
        }
    }

    // remove extra spaces and upper fist character
    const normalizedDetails = normalizeDetails(cloneProgress)

    /**
     * for the details like:
     * 1. * [new branch] master -> master
     * 2. * [new ref] refs/pull/9/merge -> refs/pull/9/merge
     * use previously parsed progress value or 0
     */
    let percentage = 0
    if (normalizedDetails.startsWith('*')) {
        percentage = previousPercentage[id] || 0
    } else {
        percentage = findProgressPercentage(cloneProgress)
        previousPercentage[id] = percentage
    }

    return {
        cloned,
        details: normalizedDetails,
        progress: percentage,
    }
}

const findProgressPercentage = (progress: string = ''): number => {
    // fist 1-3 digits before the % sign
    const PERCENTAGE = /\d{1,3}(?=%)/
    const match = progress.match(PERCENTAGE)

    if (match) {
        // convert first string into numbers
        return +match[0]
    }

    return 0
}

const normalizeDetails = (string: string): string => upperFirst(string.replace(/\s\s+/g, ' '))
const ignoreExternalService = (fullRepoName: string): string => fullRepoName.slice(fullRepoName.indexOf('/') + 1)
