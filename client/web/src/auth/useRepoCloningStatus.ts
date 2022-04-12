import { ApolloError, ReactiveVar } from '@apollo/client'
import { upperFirst, xor } from 'lodash'

import { useQuery, gql } from '@sourcegraph/http-client'

import { Maybe, UserRepositoriesVariables } from '../graphql-operations'

import { RepoSelectionMode } from './PostSignUpPage'
import { useAffiliatedRepos } from './useAffiliatedRepos'
import { MinSelectedRepo } from './useSelectedRepos'

interface UseRepoCloningStatusArguments {
    userId: string
    pollInterval: number
    selectedReposVar: ReactiveVar<MinSelectedRepo[] | undefined>
    repoSelectionMode: RepoSelectionMode
}

export interface RepoLine {
    id: string
    title: string
    details: string
    progress: number
}

export interface RepoCloningStatus {
    repos: RepoLine[] | undefined
    loading: boolean
    isDoneCloning: boolean
    error: ApolloError | undefined
    statusSummary: string
    stopPolling: () => void
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

interface ParsedMirrorInfo {
    progress: number
    details: string
    cloned: boolean
}

// temp object to store previous cloning progress
let previousMirrorInfo: { [key: string]: ParsedMirrorInfo } = {}

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
                    totalCount
                }
            }
        }
    }
`

const getRepoNames = (repos: { name: string }[] | undefined): string[] | [] =>
    repos ? repos.map(({ name }) => name) : []

export const useRepoCloningStatus = ({
    userId,
    pollInterval = 5000,
    selectedReposVar,
    repoSelectionMode,
}: UseRepoCloningStatusArguments): RepoCloningStatus => {
    let clonedReposCount = 0

    const { called, data, loading, error, stopPolling } = useQuery<CloneProgressResult, UserRepositoriesVariables>(
        USER_AFFILIATED_REPOS_MIRROR_INFO,
        {
            variables: {
                id: userId,
                cloned: true,
                notCloned: true,
                indexed: true,
                notIndexed: true,
                first: 2000,
                query: null,
                externalServiceID: null,
            },
            pollInterval,
            fetchPolicy: 'no-cache',
        }
    )

    let didReceiveAllRepoStatuses = false
    const repos = data?.node?.repositories.nodes
    const isSyncingAllRepos = repoSelectionMode === 'all'
    const { affiliatedRepos } = useAffiliatedRepos(userId)
    // check if we received cloning status for all selected repos
    const selectedRepos = selectedReposVar()

    if (isSyncingAllRepos && affiliatedRepos && repos) {
        didReceiveAllRepoStatuses = affiliatedRepos.length === repos.length
    } else {
        didReceiveAllRepoStatuses = xor(getRepoNames(selectedRepos), getRepoNames(repos)).length === 0
    }

    // don't display repo statuses unless we received all repos
    // when user goes back to reselect repos and navigates to the terminal UI
    // the endpoint may still respond with statuses for the previous selection
    if (!Array.isArray(repos) || !didReceiveAllRepoStatuses) {
        return {
            repos: undefined,
            isDoneCloning: false,
            loading,
            error,
            stopPolling,
            statusSummary: '',
        }
    }

    const repoLines: RepoLine[] = repos.reduce((lines, { id, name, mirrorInfo }) => {
        const { details, progress, cloned } = parseMirrorInfo(id, mirrorInfo)

        if (cloned) {
            clonedReposCount++
        }

        lines.push({
            id,
            details,
            progress,
            title: ignoreExternalService(name),
        })

        return lines
    }, [] as RepoLine[])

    // sort alphabetically
    repoLines.sort((lineA, lineB) => {
        const nameA = lineA.title.toUpperCase()
        const nameB = lineB.title.toUpperCase()

        return nameA < nameB ? -1 : nameA > nameB ? 1 : 0
    })

    const isDoneCloning = didReceiveAllRepoStatuses && clonedReposCount === repoLines.length

    // stop polling and cleanup memory when all repos are done cloning
    if (called && isDoneCloning) {
        stopPolling()
        previousMirrorInfo = {}
    }

    return {
        repos: repoLines,
        isDoneCloning,
        loading,
        error,
        stopPolling,
        statusSummary: `${clonedReposCount}/${repoLines.length} repositories synced`,
    }
}

const parseMirrorInfo = (id: string, mirrorInfo: RepoFields['mirrorInfo']): ParsedMirrorInfo => {
    const { cloneProgress, cloned, cloneInProgress } = mirrorInfo

    // cloned
    if (cloned) {
        return {
            cloned,
            details: 'Successfully cloned',
            progress: 100,
        }
    }

    if (!cloneInProgress) {
        // If we have previous mirror info - return it.
        // Endpoint may return that cloning is not in progress because of async
        // behavior on the backend. But if it was started we should continue
        // showing repo cloning as "in progress" on the FE
        if (previousMirrorInfo[id]) {
            return previousMirrorInfo[id]
        }

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
        const reposPreviousMirrorInfo = previousMirrorInfo[id]
        percentage = reposPreviousMirrorInfo?.progress || 0
    } else {
        percentage = findProgressPercentage(cloneProgress)
    }

    const parsedMirrorInfo = {
        cloned,
        details: normalizedDetails,
        progress: percentage,
    }

    if (percentage !== 0) {
        previousMirrorInfo[id] = parsedMirrorInfo
    }

    return parsedMirrorInfo
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
