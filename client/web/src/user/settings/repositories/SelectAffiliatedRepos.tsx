import React, { useCallback, useEffect, useState, FunctionComponent, Dispatch, SetStateAction } from 'react'

import { FetchResult } from '@apollo/client'
import classNames from 'classnames'
import { isEqual } from 'lodash'

import { ErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageSelector, RadioButton, Select, Link, Checkbox, Text } from '@sourcegraph/wildcard'

import { RepoSelectionMode } from '../../../auth/PostSignUpPage'
import { useSteps } from '../../../auth/Steps'
import { useAffiliatedRepos } from '../../../auth/useAffiliatedRepos'
import { useExternalServices } from '../../../auth/useExternalServices'
import { useSelectedRepos, selectedReposVar, MinSelectedRepo } from '../../../auth/useSelectedRepos'
import { AwayPrompt } from '../../../components/AwayPrompt'
import {
    ExternalServiceKind,
    Maybe,
    SiteAdminRepositoryFields,
    SetExternalServiceReposResult,
} from '../../../graphql-operations'

import {
    FilterInput,
    ListItemContainer,
    RepositoryNodeContainer,
    ShimmerContainer,
    UserSettingReposContainer,
} from './components'
import { CheckboxRepositoryNode } from './RepositoryNode'

import styles from './SelectAffiliatedRepos.module.scss'

export interface AffiliatedReposReference {
    submit: () => Promise<FetchResult<SetExternalServiceReposResult>[] | void>
}

interface authenticatedUser {
    id: string
    siteAdmin: boolean
    tags: string[]
}

interface Props extends TelemetryProps {
    authenticatedUser: authenticatedUser
    onRepoSelectionModeChange: Dispatch<SetStateAction<RepoSelectionMode>>
    repoSelectionMode: RepoSelectionMode
    onError: (error: ErrorLike) => void
}

export interface Repo {
    name: string
    codeHost: Maybe<{ kind: ExternalServiceKind; id: string; displayName: string }>
    private: boolean
    mirrorInfo?: SiteAdminRepositoryFields['mirrorInfo']
}

interface GitHubConfig {
    repos?: string[]
    repositoryQuery?: string[]
    token: 'REDACTED'
    url: string
}

interface GitLabConfig {
    projectQuery?: string[]
    projects?: { name: string }[]
    token: 'REDACTED'
    url: string
}

const PER_PAGE = 20

// project queries that are used when user syncs all repos from a code host
const GITLAB_SYNC_ALL_PROJECT_QUERY = 'projects?membership=true&archived=no'
const GITHUB_SYNC_ALL_PROJECT_QUERY = 'affiliated'

// initial state constants
const emptyRepos: Repo[] = []
const initialRepoState = {
    repos: emptyRepos,
    loading: false,
    loaded: false,
}

const initialSelectionState = {
    repos: new Map<string, Repo>(),
    loaded: false,
    radio: '',
}

/**
 * A page to manage the repositories a user syncs from their connected code hosts.
 */
export const SelectAffiliatedRepos: FunctionComponent<React.PropsWithChildren<Props>> = ({
    authenticatedUser,
    onRepoSelectionModeChange,
    repoSelectionMode,
    telemetryService,
    onError,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    const { setComplete, resetToTheRight, currentIndex, setStep } = useSteps()
    const { externalServices, errorServices } = useExternalServices(authenticatedUser.id)
    const { affiliatedRepos, errorAffiliatedRepos } = useAffiliatedRepos(authenticatedUser.id)
    const { selectedRepos, errorSelectedRepos } = useSelectedRepos(authenticatedUser.id)

    const fetchingError =
        errorServices ||
        // The affliliated repos query will always return an error on the GraphQL API when no
        // external services are set up. In this case we allow the user to go back to the previous
        // step and set up code hosts.
        (externalServices !== undefined && externalServices.length !== 0 && errorAffiliatedRepos) ||
        errorSelectedRepos

    useEffect(() => {
        if (fetchingError) {
            onError(fetchingError)
        }
    }, [fetchingError, onError])

    // set up state hooks
    const [currentPage, setPage] = useState(1)
    const [repoState, setRepoState] = useState(initialRepoState)
    const [onloadSelectedRepos, setOnloadSelectedRepos] = useState<string[]>([])
    const [selectionState, setSelectionState] = useState(initialSelectionState)
    const [didSelectionChange, setDidSelectionChange] = useState(false)
    const [query, setQuery] = useState('')
    const [codeHostFilter, setCodeHostFilter] = useState('')
    const [filteredRepos, setFilteredRepos] = useState<Repo[]>([])

    const getRepoServiceAndName = (repo: Repo): string => `${repo.codeHost?.kind || 'unknown'}/${repo.name}`

    useEffect(() => {
        if (externalServices && affiliatedRepos) {
            const codeHostsHaveSyncAllQuery = []

            for (const host of externalServices) {
                if (host.lastSyncError || host.warning) {
                    continue
                }

                const cfg = JSON.parse(host.config) as GitHubConfig | GitLabConfig
                switch (host.kind) {
                    case ExternalServiceKind.GITLAB: {
                        const gitLabCfg = cfg as GitLabConfig

                        if (Array.isArray(gitLabCfg.projectQuery)) {
                            codeHostsHaveSyncAllQuery.push(
                                gitLabCfg.projectQuery.includes(GITLAB_SYNC_ALL_PROJECT_QUERY)
                            )
                        }

                        break
                    }

                    case ExternalServiceKind.GITHUB: {
                        const gitHubCfg = cfg as GitHubConfig

                        if (Array.isArray(gitHubCfg.repositoryQuery)) {
                            codeHostsHaveSyncAllQuery.push(
                                gitHubCfg.repositoryQuery.includes(GITHUB_SYNC_ALL_PROJECT_QUERY)
                            )
                        }

                        break
                    }
                }
            }

            const selectedAffiliatedRepos = new Map<string, Repo>()

            const cachedSelectedRepos = selectedReposVar()
            const userSelectedRepos = repoSelectionMode === 'all' ? [] : cachedSelectedRepos || selectedRepos || []

            const affiliatedReposWithMirrorInfo = affiliatedRepos.map(affiliatedRepo => {
                let foundInSelected: SiteAdminRepositoryFields | MinSelectedRepo | null = null
                for (const selectedRepo of userSelectedRepos) {
                    const {
                        name,
                        externalRepository: { serviceType: selectedRepoServiceType },
                    } = selectedRepo
                    const selectedRepoName = name.slice(name.indexOf('/') + 1)

                    if (
                        selectedRepoName === affiliatedRepo.name &&
                        selectedRepoServiceType === affiliatedRepo.codeHost?.kind.toLocaleLowerCase()
                    ) {
                        foundInSelected = selectedRepo
                        break
                    }
                }

                if (foundInSelected) {
                    // save off only selected repos
                    selectedAffiliatedRepos.set(getRepoServiceAndName(affiliatedRepo), affiliatedRepo)
                }

                return affiliatedRepo
            })

            // sort affiliated repos with already selected repos at the top
            affiliatedReposWithMirrorInfo.sort((repoA, repoB): number => {
                const isRepoASelected = selectedAffiliatedRepos.has(getRepoServiceAndName(repoA))
                const isRepoBSelected = selectedAffiliatedRepos.has(getRepoServiceAndName(repoB))

                if (!isRepoASelected && isRepoBSelected) {
                    return 1
                }

                if (isRepoASelected && !isRepoBSelected) {
                    return -1
                }

                return 0
            })

            // safe off initial selection state
            setOnloadSelectedRepos(previousValue => [...previousValue, ...selectedAffiliatedRepos.keys()])

            /**
             * 1. if every code host has a project query to sync all repos or the
             * number of affiliated repos equals to the number of selected repos -
             * set radio to 'all'
             * 2. if only some repos were selected - set radio to 'selected'
             * 3. if no repos selected - empty state
             */

            const radioSelectOption =
                repoSelectionMode ||
                (externalServices.length === codeHostsHaveSyncAllQuery.length &&
                codeHostsHaveSyncAllQuery.every(Boolean)
                    ? 'all'
                    : selectedAffiliatedRepos.size > 0
                    ? 'selected'
                    : '')

            onRepoSelectionModeChange(radioSelectOption as RepoSelectionMode)

            // set sorted repos and mark as loaded
            setRepoState(previousRepoState => ({
                ...previousRepoState,
                repos: affiliatedReposWithMirrorInfo,
                loaded: true,
            }))

            setSelectionState({
                repos: selectedAffiliatedRepos,
                radio: radioSelectOption,
                loaded: true,
            })
        }
    }, [
        externalServices,
        affiliatedRepos,
        selectedRepos,
        onRepoSelectionModeChange,
        setComplete,
        currentIndex,
        repoSelectionMode,
    ])

    // select repos by code host and query
    useEffect(() => {
        // filter our set of repos based on query & code host selection
        const filtered: Repo[] = []

        for (const repo of repoState.repos) {
            // filtering by code hosts
            if (codeHostFilter !== '' && repo.codeHost?.id !== codeHostFilter) {
                continue
            }

            const queryLoweCase = query.toLowerCase()
            const nameLowerCase = repo.name.toLowerCase()
            if (!nameLowerCase.includes(queryLoweCase)) {
                continue
            }

            filtered.push(repo)
        }

        // set new filtered pages and reset the pagination
        setFilteredRepos(filtered)
        setPage(1)
    }, [repoState.repos, codeHostFilter, query])

    const handleRadioSelect = (changeEvent: React.ChangeEvent<HTMLInputElement>): void => {
        setSelectionState({
            repos: selectionState.repos,
            radio: changeEvent.currentTarget.value,
            loaded: selectionState.loaded,
        })

        onRepoSelectionModeChange(changeEvent.currentTarget.value as RepoSelectionMode)
    }

    // calculate if the current step is completed based on repo selection when
    // we toggle between "Sync all" and individual repo selection checkboxes
    useEffect(() => {
        if (selectionState.radio) {
            if (selectionState.radio === 'all') {
                setComplete(currentIndex, true)
            } else {
                const hasSelectedRepos = selectionState.repos.size !== 0
                if (hasSelectedRepos) {
                    setComplete(currentIndex, true)
                } else {
                    setComplete(currentIndex, false)
                    resetToTheRight(currentIndex)
                }
            }
        }
    }, [currentIndex, resetToTheRight, selectionState.radio, selectionState.repos.size, setComplete])

    const hasCodeHosts = Array.isArray(externalServices) && externalServices.length > 0

    const modeSelect: JSX.Element = (
        <>
            <div className="d-flex flex-row align-items-baseline">
                <RadioButton
                    id="sync_all_repositories"
                    name="all_repositories"
                    value="all"
                    checked={selectionState.radio === 'all'}
                    onChange={handleRadioSelect}
                    label={
                        <div className="d-flex flex-column ml-2">
                            <Text className="mb-0">Sync all repositories</Text>
                            <Text weight="regular" className="text-muted">
                                Will sync all current and future public and private repositories
                            </Text>
                        </div>
                    }
                />
            </div>
            <div className="d-flex flex-row align-items-baseline mb-0">
                <RadioButton
                    id="sync_selected_repositories"
                    name="selected_repositories"
                    value="selected"
                    checked={selectionState.radio === 'selected'}
                    onChange={handleRadioSelect}
                    label={
                        <div className="d-flex flex-column ml-2">
                            <Text className="mb-0">Sync selected repositories</Text>
                        </div>
                    }
                />
            </div>
        </>
    )

    const filterControls: JSX.Element = (
        <div className="w-100 d-inline-flex justify-content-between flex-row mt-3">
            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                <Text className="text-xl-center text-nowrap mr-2">Code Host:</Text>
                <Select
                    name="code-host"
                    aria-label="select code host type"
                    onChange={event => setCodeHostFilter(event.target.value)}
                >
                    <option key="any" value="" label="Any" />
                    {externalServices?.map(value => (
                        <option key={value.id} value={value.id} label={value.displayName} />
                    ))}
                </Select>
            </div>
            <FilterInput
                type="search"
                placeholder="Filter..."
                name="query"
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                onChange={event => {
                    setQuery(event.target.value)
                }}
            />
        </div>
    )

    const saveRepoSelection = (repos: Repo[]): void => {
        // save off last selected repos
        const selection = repos.reduce((accumulator, repo) => {
            const serviceType = repo.codeHost?.kind.toLowerCase()
            const serviceName = serviceType ? `${serviceType}.com` : 'unknown'

            accumulator.push({
                name: `${serviceName}/${repo.name}`,
                externalRepository: { serviceType: serviceType || 'unknown', id: repo.codeHost?.id },
            })
            return accumulator
        }, [] as MinSelectedRepo[])

        // safe off repo selection to apollo
        selectedReposVar(selection)
    }

    const onRepoClicked = useCallback(
        (repo: Repo) => (): void => {
            const clickedRepo = getRepoServiceAndName(repo)
            const newSelection = new Map(selectionState.repos)
            if (newSelection.has(clickedRepo)) {
                newSelection.delete(clickedRepo)
            } else {
                newSelection.set(clickedRepo, repo)
            }

            const currentlySelectedRepos = [...newSelection.keys()]
            const didChange = !isEqual(currentlySelectedRepos.sort(), onloadSelectedRepos.sort())

            // set current step as complete if user had already selected repos or made changes
            if (didChange || onloadSelectedRepos.length !== 0) {
                setComplete(currentIndex, true)
                setDidSelectionChange(true)
            } else {
                setComplete(currentIndex, false)
                resetToTheRight(currentIndex)
            }

            saveRepoSelection([...newSelection.values()])

            // set new selection state
            setSelectionState({
                repos: newSelection,
                radio: selectionState.radio,
                loaded: selectionState.loaded,
            })
        },
        [
            currentIndex,
            onloadSelectedRepos,
            resetToTheRight,
            selectionState.loaded,
            selectionState.radio,
            selectionState.repos,
            setComplete,
        ]
    )

    const getSelectedReposByCodeHost = (codeHostId: string = ''): Repo[] => {
        const selectedRepos = [...selectionState.repos.values()]
        // if no specific code host selected, return all selected repos
        return codeHostId ? selectedRepos.filter(({ codeHost }) => codeHost?.id === codeHostId) : selectedRepos
    }

    const areAllReposSelected = (): boolean => {
        if (selectionState.repos.size === 0) {
            return false
        }

        const selectedRepos = getSelectedReposByCodeHost(codeHostFilter)
        return selectedRepos.length === filteredRepos.length
    }

    const selectAll = (): void => {
        const newSelection = new Map<string, Repo>()
        // if not all repos are selected, we should select all, otherwise empty the selection

        if (selectionState.repos.size !== filteredRepos.length) {
            for (const repo of filteredRepos) {
                newSelection.set(getRepoServiceAndName(repo), repo)
            }
        }

        saveRepoSelection([...newSelection.values()])

        setSelectionState({
            repos: newSelection,
            loaded: selectionState.loaded,
            radio: selectionState.radio,
        })
    }

    const rows: JSX.Element = (
        <tbody>
            <tr className="align-items-baseline d-flex" key="header">
                <RepositoryNodeContainer
                    as="td"
                    className="p-2 w-100 d-flex align-items-center border-top-0 border-bottom"
                >
                    <Checkbox
                        id="select-all-repos"
                        className="mr-3"
                        checked={areAllReposSelected()}
                        onChange={selectAll}
                        label={
                            <small
                                className={classNames({
                                    'text-muted': selectionState.repos.size === 0,
                                    'text-body': selectionState.repos.size !== 0,
                                    'mb-0': true,
                                })}
                            >
                                {selectionState.repos.size === 0
                                    ? 'Select all'
                                    : `${selectionState.repos.size} ${
                                          selectionState.repos.size === 1 ? 'repository' : 'repositories'
                                      } selected`}
                            </small>
                        }
                    />
                </RepositoryNodeContainer>
            </tr>
            {filteredRepos.map((repo, index) => {
                if (index < (currentPage - 1) * PER_PAGE || index >= currentPage * PER_PAGE) {
                    return
                }

                const serviceAndRepoName = getRepoServiceAndName(repo)

                return (
                    <CheckboxRepositoryNode
                        name={repo.name}
                        key={serviceAndRepoName}
                        onClick={onRepoClicked(repo)}
                        checked={selectionState.repos.has(serviceAndRepoName)}
                        serviceType={repo.codeHost?.kind || 'unknown'}
                        isPrivate={repo.private}
                    />
                )
            })}
        </tbody>
    )

    const modeSelectShimmer: JSX.Element = (
        <div className="container">
            <div className="mt-2 row">
                <ShimmerContainer circle={true} className="mr-2" />
                <ShimmerContainer className="mb-1 p-2 border-top-0 col-sm-2" />
            </div>
            <div className="mt-1 ml-2 row">
                <ShimmerContainer className="mb-3 p-2 ml-1 border-top-0 col-sm-6" />
            </div>
            <div className="mt-2 row">
                <ShimmerContainer circle={true} className="mr-2" />
                <ShimmerContainer className="p-2 mb-1 border-top-0 col-sm-3" />
            </div>
        </div>
    )

    return (
        <UserSettingReposContainer className="mb-0">
            <Container>
                <ul className="list-group">
                    <ListItemContainer key="from-code-hosts">
                        {externalServices && !hasCodeHosts ? (
                            <div className={styles.noCodeHosts}>
                                <Text>
                                    <Link
                                        to="/welcome"
                                        onClick={event => {
                                            event.preventDefault()
                                            event.stopPropagation()
                                            setStep(Math.max(0, currentIndex - 1))
                                        }}
                                    >
                                        Add a code host
                                    </Link>{' '}
                                    to add repositories.
                                </Text>
                            </div>
                        ) : (
                            <div>
                                {/* display type of repo sync radio buttons or shimmer when appropriate */}
                                {hasCodeHosts && selectionState.loaded ? modeSelect : modeSelectShimmer}

                                {hasCodeHosts && selectionState.radio === 'selected' && (
                                    <div className="ml-4">
                                        {filterControls}
                                        <table role="grid" className="table">
                                            {
                                                // if the repos are loaded display the rows of repos
                                                repoState.loaded && rows
                                            }
                                        </table>
                                        {filteredRepos.length > 0 && (
                                            <PageSelector
                                                currentPage={currentPage}
                                                onPageChange={setPage}
                                                totalPages={Math.ceil(filteredRepos.length / PER_PAGE)}
                                                className="pt-4"
                                            />
                                        )}
                                    </div>
                                )}
                            </div>
                        )}
                    </ListItemContainer>
                </ul>
            </Container>
            <AwayPrompt
                header="Discard unsaved changes?"
                message="Currently synced repositories will be unchanged"
                button_ok_text="Discard"
                when={didSelectionChange}
            />
        </UserSettingReposContainer>
    )
}
