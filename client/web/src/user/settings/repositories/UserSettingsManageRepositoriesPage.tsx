import React, { ChangeEvent, FormEvent, useCallback, useEffect, useState, useRef } from 'react'

import classNames from 'classnames'
import { isEqual, capitalize } from 'lodash'
import { useHistory } from 'react-router'
import { Subscription } from 'rxjs'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    ProductStatusBadge,
    Container,
    PageSelector,
    RadioButton,
    TextArea,
    Select,
    Button,
    Alert,
    Link,
    Checkbox,
    H2,
    H3,
    H4,
    Text,
} from '@sourcegraph/wildcard'

import { ALLOW_NAVIGATION, AwayPrompt } from '../../../components/AwayPrompt'
import {
    queryExternalServices,
    setExternalServiceRepos,
    listAffiliatedRepositories,
} from '../../../components/externalServices/backend'
import { LoaderButton } from '../../../components/LoaderButton'
import { PageTitle } from '../../../components/PageTitle'
import {
    ExternalServiceKind,
    ExternalServicesResult,
    Maybe,
    AffiliatedRepositoriesResult,
    RepositoriesResult,
    SiteAdminRepositoryFields,
} from '../../../graphql-operations'
import {
    listUserRepositories,
    queryUserPublicRepositories,
    setUserPublicRepositories,
} from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { externalServiceUserModeFromTags, Owner } from '../cloud-ga'

import {
    FilterInput,
    ListItemContainer,
    RepositoryNodeContainer,
    ShimmerContainer,
    UserSettingReposContainer,
} from './components'
import { CheckboxRepositoryNode } from './RepositoryNode'

import styles from './UserSettingsManageRepositoriesPage.module.scss'

interface Props extends TelemetryProps {
    owner: Owner
    routingPrefix: string
}

interface Repo {
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

const PER_PAGE = 25

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

const emptyHosts: ExternalServicesResult['externalServices']['nodes'] = []

const initialCodeHostState = {
    hosts: emptyHosts,
    loaded: false,
}
const initialPublicRepoState = {
    repos: '',
    enabled: false,
    loaded: false,
}
const initialSelectionState = {
    repos: new Map<string, Repo>(),
    loaded: false,
    radio: '',
}

type initialFetchingReposState = undefined | 'loading'
type affiliateRepoProblemType = undefined | string | ErrorLike | ErrorLike[]

const displayWarning = (warning: string, hint?: JSX.Element): JSX.Element => (
    <Alert className="my-3" role="alert" key={warning} variant="warning">
        <H4 className="align-middle mb-1">{capitalize(warning)}</H4>
        <Text className="align-middle mb-0">
            {hint} {hint ? 'for more details.' : null}
        </Text>
    </Alert>
)

const displayError = (error: ErrorLike, hint?: JSX.Element): JSX.Element => (
    <Alert className="my-3" role="alert" key={error.message} variant="danger">
        <H4 className="align-middle mb-1">{capitalize(error.message)}</H4>
        <Text className="align-middle mb-0">
            {hint} {hint ? 'for more details.' : null}
        </Text>
    </Alert>
)

const displayAffiliateRepoProblems = (
    problem: affiliateRepoProblemType,
    hint?: JSX.Element
): JSX.Element | JSX.Element[] | null => {
    if (typeof problem === 'string') {
        return displayWarning(problem, hint)
    }

    if (isErrorLike(problem)) {
        return displayError(problem, hint)
    }

    if (Array.isArray(problem)) {
        return <>{problem.map(prob => displayAffiliateRepoProblems(prob, hint))}</>
    }

    return null
}

/**
 * A page to manage the repositories a user syncs from their connected code hosts.
 */
export const UserSettingsManageRepositoriesPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    owner,
    routingPrefix,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    const history = useHistory()
    const isOrgOwner = false

    // if we should tweak UI messaging and copy
    const ALLOW_PRIVATE_CODE = externalServiceUserModeFromTags() === 'all'

    // set up state hooks
    const [repoState, setRepoState] = useState(initialRepoState)
    const [publicRepoState, setPublicRepoState] = useState(initialPublicRepoState)
    const [codeHosts, setCodeHosts] = useState(initialCodeHostState)
    const [onloadSelectedRepos, setOnloadSelectedRepos] = useState<string[]>([])
    const [onloadRadioValue, setOnloadRadioValue] = useState('')
    const [selectionState, setSelectionState] = useState(initialSelectionState)
    const [currentPage, setPage] = useState(1)
    const [query, setQuery] = useState(new URLSearchParams(window.location.search).get('filter') || '')
    const [codeHostFilter, setCodeHostFilter] = useState('')
    const [filteredRepos, setFilteredRepos] = useState<Repo[]>([])
    const [fetchingRepos, setFetchingRepos] = useState<initialFetchingReposState>()
    const externalServiceSubscription = useRef<Subscription>()

    // since we're making many different GraphQL requests - track affiliate and
    // manually added public repo errors separately
    const [affiliateRepoProblems, setAffiliateRepoProblems] = useState<affiliateRepoProblemType>()
    const [otherPublicRepoError, setOtherPublicRepoError] = useState<undefined | ErrorLike>()

    const ExternalServiceProblemHint = (
        <Link className="font-weight-normal" to={`${routingPrefix}/code-hosts`}>
            Check code host connections
        </Link>
    )

    const toggleTextArea = useCallback(
        () => setPublicRepoState({ ...publicRepoState, enabled: !publicRepoState.enabled }),
        [publicRepoState]
    )

    const fetchExternalServices = useCallback(
        async (): Promise<ExternalServicesResult['externalServices']['nodes']> =>
            queryExternalServices({
                first: null,
                after: null,
                namespace: owner.id,
            })
                .toPromise()
                .then(({ nodes }) => nodes),

        [owner.id]
    )

    const fetchAffiliatedRepos = useCallback(
        async (): Promise<AffiliatedRepositoriesResult['affiliatedRepositories']> =>
            listAffiliatedRepositories({
                namespace: owner.id,
                codeHost: null,
                query: null,
            })
                .toPromise()
                .then(data => data.affiliatedRepositories),

        [owner.id]
    )

    const fetchSelectedRepositories = useCallback(
        async (): Promise<NonNullable<RepositoriesResult>['repositories']['nodes']> =>
            listUserRepositories({ id: owner.id, first: 2000 })
                .toPromise()
                .then(({ nodes }) => nodes),
        [owner.id]
    )

    const getRepoServiceAndName = (repo: Repo): string => `${repo.codeHost?.kind || 'unknown'}/${repo.name}`

    const fetchServicesAndAffiliatedRepos = useCallback(async (): Promise<void> => {
        const externalServices = await fetchExternalServices()

        // short-circuit if user doesn't code hosts added
        if (externalServices.length === 0) {
            setCodeHosts({
                loaded: true,
                hosts: [],
            })

            return
        }

        // loaded code hosts
        setCodeHosts({
            loaded: true,
            hosts: externalServices,
        })

        const codeHostsHaveSyncAllQuery = []

        // if external services return errors or warnings, we can display the errors from one code host along with the repos from another.
        const codeHostProblems = []

        for (const host of externalServices) {
            let hostHasProblems = false

            if (host.lastSyncError) {
                hostHasProblems = true
                codeHostProblems.push(asError(`${host.displayName} sync error: ${host.lastSyncError}`))
            }

            if (host.warning) {
                hostHasProblems = true
                codeHostProblems.push(asError(`${host.displayName} warning: ${host.warning}`))
            }

            if (hostHasProblems) {
                // skip this code host
                continue
            }

            const cfg = JSON.parse(host.config) as GitHubConfig | GitLabConfig
            switch (host.kind) {
                case ExternalServiceKind.GITLAB: {
                    const gitLabCfg = cfg as GitLabConfig

                    if (Array.isArray(gitLabCfg.projectQuery)) {
                        codeHostsHaveSyncAllQuery.push(gitLabCfg.projectQuery.includes(GITLAB_SYNC_ALL_PROJECT_QUERY))
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

        const [affiliatedRepos, selectedRepos] = await Promise.all([
            fetchAffiliatedRepos(),
            fetchSelectedRepositories(),
        ])

        if (codeHostProblems.length > 0) {
            setAffiliateRepoProblems(codeHostProblems)
        }

        // If the external services call doen't return any errors, we can get them from the affiliated repos call.
        if (codeHostProblems.length === 0 && affiliatedRepos.codeHostErrors !== []) {
            for (const codeHostError of affiliatedRepos.codeHostErrors) {
                codeHostProblems.push(asError(codeHostError))
            }
            setAffiliateRepoProblems(codeHostProblems)
        }

        const selectedAffiliatedRepos = new Map<string, Repo>()

        const affiliatedReposWithMirrorInfo = affiliatedRepos.nodes.map(affiliatedRepo => {
            const foundInSelected = selectedRepos.find(
                ({ name, externalRepository: { serviceType: selectedRepoServiceType } }) => {
                    // selected repo names formatted: code-host/owner/repository
                    const selectedRepoName = name.slice(name.indexOf('/') + 1)

                    return (
                        selectedRepoName === affiliatedRepo.name &&
                        selectedRepoServiceType === affiliatedRepo.codeHost?.kind.toLocaleLowerCase()
                    )
                }
            )

            if (foundInSelected) {
                // save off only selected repos
                selectedAffiliatedRepos.set(getRepoServiceAndName(affiliatedRepo), affiliatedRepo)

                // add mirror info object where it exists - will be used for filters
                return { ...affiliatedRepo, mirrorInfo: foundInSelected.mirrorInfo }
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
         * 3. if no repos selected or this is an org - set radio to 'selected'
         * 4. otherwise, empty
         */

        const radioSelectOption =
            externalServices.length === codeHostsHaveSyncAllQuery.length && codeHostsHaveSyncAllQuery.every(Boolean)
                ? 'all'
                : selectedAffiliatedRepos.size > 0
                ? 'selected'
                : ''

        setOnloadRadioValue(radioSelectOption)

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
    }, [fetchExternalServices, fetchAffiliatedRepos, fetchSelectedRepositories])

    useEffect(() => {
        fetchServicesAndAffiliatedRepos().catch(error => {
            // handle different errors here
            setAffiliateRepoProblems(asError(error))
            setRepoState({
                repos: emptyRepos,
                loading: false,
                loaded: true,
            })
        })
    }, [fetchServicesAndAffiliatedRepos])

    // fetch public repos for the "other public repositories" textarea
    const fetchAndSetPublicRepos = useCallback(async (): Promise<void> => {
        const result = await queryUserPublicRepositories(owner.id).toPromise()

        if (!result) {
            setPublicRepoState({ ...initialPublicRepoState, loaded: true })
        } else {
            // public repos separated by a new line
            const publicRepos = result.map(({ name }) => name)

            // safe off initial selection state
            setOnloadSelectedRepos(previousValue => [...previousValue, ...publicRepos])

            setPublicRepoState({ repos: publicRepos.join('\n'), loaded: true, enabled: result.length > 0 })
        }
    }, [owner.id])

    useEffect(() => {
        if (!isOrgOwner) {
            fetchAndSetPublicRepos().catch(error => setOtherPublicRepoError(asError(error)))
        }
    }, [fetchAndSetPublicRepos, isOrgOwner])

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

    const didRepoSelectionChange = useCallback((): boolean => {
        const publicRepos = publicRepoState.enabled && publicRepoState.repos ? publicRepoState.repos.split('\n') : []
        const affiliatedRepos = selectionState.repos.keys()

        const currentlySelectedRepos = [...publicRepos, ...affiliatedRepos]

        return (
            selectionState.radio !== onloadRadioValue ||
            !isEqual(currentlySelectedRepos.sort(), onloadSelectedRepos.sort())
        )
    }, [
        onloadSelectedRepos,
        publicRepoState.enabled,
        publicRepoState.repos,
        selectionState.repos,
        selectionState.radio,
        onloadRadioValue,
    ])

    // save changes and update code hosts
    const submit = useCallback(
        async (event: FormEvent<HTMLFormElement>): Promise<void> => {
            event.preventDefault()

            const publicRepos = publicRepoState.enabled
                ? publicRepoState.repos.split('\n').filter((row): boolean => row !== '')
                : []

            const loggerPayload = {
                userReposSelection: selectionState.radio
                    ? selectionState.radio === 'selected'
                        ? 'specific'
                        : 'all'
                    : null,
                didAddReposByURL: !!publicRepos.length,
            }
            eventLogger.log('UserSettingsManageRepositoriesSaved', loggerPayload, loggerPayload)

            setFetchingRepos('loading')

            if (!isOrgOwner) {
                try {
                    await setUserPublicRepositories(owner.id, publicRepos).toPromise()
                } catch (error) {
                    setOtherPublicRepoError(asError(error))
                    setFetchingRepos(undefined)
                    return
                }
            }

            if (!selectionState.radio) {
                // location state is used here to prevent AwayPrompt from blocking
                return history.push(routingPrefix + '/repositories', ALLOW_NAVIGATION)
            }

            const codeHostRepoPromises = []

            for (const host of codeHosts.hosts) {
                const repos: string[] = []
                for (const repo of selectionState.repos.values()) {
                    if (repo.codeHost?.id !== host.id) {
                        continue
                    }
                    repos.push(repo.name)
                }

                codeHostRepoPromises.push(
                    setExternalServiceRepos({
                        id: host.id,
                        allRepos: selectionState.radio === 'all',
                        repos: (selectionState.radio === 'selected' && repos) || null,
                    })
                )
            }

            try {
                await Promise.all(codeHostRepoPromises)
            } catch (error) {
                setAffiliateRepoProblems(asError(error))
                setFetchingRepos(undefined)
                return
            }

            // location state is used here to prevent AwayPrompt from blocking
            return history.push(routingPrefix + '/repositories', ALLOW_NAVIGATION)
        },
        [
            publicRepoState.enabled,
            publicRepoState.repos,
            selectionState.radio,
            selectionState.repos,
            isOrgOwner,
            history,
            routingPrefix,
            owner.id,
            codeHosts.hosts,
        ]
    )

    useEffect(
        () => () => {
            externalServiceSubscription.current?.unsubscribe()
        },
        []
    )

    const handleRadioSelect = (changeEvent: React.ChangeEvent<HTMLInputElement>): void => {
        setSelectionState({
            repos: selectionState.repos,
            radio: changeEvent.currentTarget.value,
            loaded: selectionState.loaded,
        })
    }

    // code hosts were loaded and some were configured
    const hasCodeHosts = codeHosts.loaded && codeHosts.hosts.length !== 0

    const modeSelect: JSX.Element = (
        <Form className="mt-4">
            {!isOrgOwner && (
                <div className="d-flex flex-row align-items-baseline">
                    <RadioButton
                        name="all_repositories"
                        id="sync_all_repositories"
                        value="all"
                        disabled={!hasCodeHosts}
                        checked={selectionState.radio === 'all'}
                        onChange={handleRadioSelect}
                        label={
                            <div className="d-flex flex-column ml-2">
                                <Text className="mb-0">Sync all repositories</Text>
                                <Text className="font-weight-normal text-muted">
                                    Will sync all current and future public and private repositories
                                </Text>
                            </div>
                        }
                    />
                </div>
            )}

            <div className="d-flex flex-row align-items-baseline mb-0">
                <RadioButton
                    name="selected_repositories"
                    id="sync_selected_repositories"
                    value="selected"
                    checked={selectionState.radio === 'selected'}
                    disabled={!hasCodeHosts}
                    onChange={handleRadioSelect}
                    label={
                        <div className="d-flex flex-column ml-2">
                            <Text className={classNames('mb-0', !hasCodeHosts && styles.textDisabled)}>
                                Sync selected repositories
                            </Text>
                        </div>
                    }
                />
            </div>
        </Form>
    )

    const preventSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => event.preventDefault(), [])

    const filterControls: JSX.Element = (
        <Form onSubmit={preventSubmit} className="w-100 d-inline-flex justify-content-between flex-row mt-3">
            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                <Text className="text-xl-center text-nowrap mr-2">Code Host:</Text>
                <Select
                    name="code-host"
                    aria-label="select code host type"
                    onChange={event => setCodeHostFilter(event.target.value)}
                >
                    <option key="all" value="" label="All" />
                    {codeHosts.hosts.map(value => (
                        <option key={value.id} value={value.id} label={value.displayName} />
                    ))}
                </Select>
            </div>
            <FilterInput
                type="search"
                placeholder="Filter repositories..."
                name="query"
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
                defaultValue={query}
                onChange={event => {
                    setQuery(event.target.value)
                }}
            />
        </Form>
    )

    const onRepoClicked = useCallback(
        (repo: Repo) => (): void => {
            const clickedRepo = getRepoServiceAndName(repo)
            const newMap = new Map(selectionState.repos)
            if (newMap.has(clickedRepo)) {
                newMap.delete(clickedRepo)
            } else {
                newMap.set(clickedRepo, repo)
            }
            setSelectionState({
                repos: newMap,
                radio: selectionState.radio,
                loaded: selectionState.loaded,
            })
        },
        [selectionState, setSelectionState]
    )

    const areAllFilteredReposSelected = useCallback((): boolean => {
        if (selectionState.repos.size === 0 || filteredRepos.length === 0) {
            return false
        }

        // if selection state does not contain all of the filtered repos, return false
        return !filteredRepos.some(repo => !selectionState.repos.has(getRepoServiceAndName(repo)))
    }, [selectionState, filteredRepos])

    const toggleAll = (event: ChangeEvent<HTMLInputElement>): void => {
        const { checked } = event.target
        const newSelectAll = new Map<string, Repo>(selectionState.repos)

        for (const repo of filteredRepos) {
            // if checkbox is checked, we should add filtered repo, otherwise we remove
            if (checked) {
                newSelectAll.set(getRepoServiceAndName(repo), repo)
            } else {
                newSelectAll.delete(getRepoServiceAndName(repo))
            }
        }
        setSelectionState({
            repos: newSelectAll,
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
                        checked={areAllFilteredReposSelected()}
                        onChange={toggleAll}
                        disabled={filteredRepos.length === 0}
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

    const handlePublicReposChanged = (event: React.ChangeEvent<HTMLTextAreaElement>): void => {
        setPublicRepoState({ ...publicRepoState, repos: event.target.value })
    }

    const modeSelectShimmer: JSX.Element = (
        <div className="container mt-4">
            {!isOrgOwner && (
                <>
                    <div className="mt-2 row">
                        <ShimmerContainer circle={true} className="mr-2" />
                        <ShimmerContainer className="mb-1 p-2 border-top-0 col-sm-2" />
                    </div>
                    <div className="mt-1 ml-2 row">
                        <ShimmerContainer className="mb-3 p-2 ml-1 border-top-0 col-sm-6" />
                    </div>
                </>
            )}

            <div className="mt-2 row">
                <ShimmerContainer circle={true} className="mr-2" />
                <ShimmerContainer className="p-2 mb-1 border-top-0 col-sm-3" />
            </div>
        </div>
    )

    return (
        <UserSettingReposContainer>
            <PageTitle title="Manage Repositories" />
            <H2 className="d-flex mb-2">
                Manage Repositories <ProductStatusBadge status="beta" className="ml-2" linkToDocs={true} />
            </H2>
            <Text className="text-muted">
                Choose repositories to sync with Sourcegraph.
                <Link
                    to="https://docs.sourcegraph.com/code_search/how-to/adding_repositories_to_cloud"
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    {' '}
                    Learn more about who can see code on Sourcegraph
                </Link>
                .
            </Text>
            <Container>
                <ul className="list-group">
                    <ListItemContainer key="from-code-hosts">
                        <div>
                            <H3>{owner.name ? `${owner.name}'s` : 'Your'} repositories</H3>

                            <Text className="text-muted">
                                Repositories{' '}
                                {isOrgOwner ? 'that can be synced through' : 'you own or collaborate on from your'}{' '}
                                <Link to={`${routingPrefix}/code-hosts`}>connected code hosts</Link>
                            </Text>

                            {!ALLOW_PRIVATE_CODE && hasCodeHosts && (
                                <Alert variant="primary">
                                    Coming soon: search private repositories with Sourcegraph Cloud.{' '}
                                    <Link
                                        to="https://share.hsforms.com/1copeCYh-R8uVYGCpq3s4nw1n7ku"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        Get updated when this feature launches
                                    </Link>
                                </Alert>
                            )}
                            {codeHosts.loaded && codeHosts.hosts.length === 0 && (
                                <Alert className="mb-2" variant="warning">
                                    <Link className="font-weight-normal" to={`${routingPrefix}/code-hosts`}>
                                        Connect with a code host
                                    </Link>{' '}
                                    to add
                                    {owner.name ? ` ${owner.name}'s` : ' your own'} repositories to Sourcegraph.
                                </Alert>
                            )}
                            {displayAffiliateRepoProblems(affiliateRepoProblems, ExternalServiceProblemHint)}

                            {/* display radio buttons shimmer only when user has code hosts */}
                            {hasCodeHosts && !selectionState.loaded && modeSelectShimmer}

                            {/* display type of repo sync radio buttons */}
                            {hasCodeHosts && selectionState.loaded && modeSelect}

                            {
                                // if we're in 'selected' mode, show a list of all the repos on the code hosts to select from
                                hasCodeHosts && selectionState.radio === 'selected' && (
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
                                )
                            }
                        </div>
                    </ListItemContainer>
                    {window.context.sourcegraphDotComMode && !isOrgOwner && (
                        <ListItemContainer key="add-textarea">
                            <div>
                                <H3>Other public repositories</H3>
                                <Text className="text-muted">Public repositories on GitHub and GitLab</Text>
                                <Checkbox
                                    id="add-public-repos"
                                    className="mr-2 mt-0"
                                    wrapperClassName="d-flex align-items-center"
                                    label="Sync specific public repositories by URL"
                                    onChange={toggleTextArea}
                                    checked={publicRepoState.enabled}
                                />

                                {publicRepoState.enabled && (
                                    <div className="form-group ml-4 mt-3">
                                        <Text className="mb-2">Repositories to sync</Text>
                                        <TextArea
                                            rows={5}
                                            value={publicRepoState.repos}
                                            onChange={handlePublicReposChanged}
                                        />
                                        <Text className="text-muted mt-2">
                                            Specify with complete URLs. One repository per line.
                                        </Text>
                                    </div>
                                )}
                            </div>
                        </ListItemContainer>
                    )}
                </ul>
            </Container>
            {isErrorLike(otherPublicRepoError) && displayError(otherPublicRepoError)}
            <AwayPrompt
                header="Discard unsaved changes?"
                message="Currently synced repositories will be unchanged"
                button_ok_text="Discard"
                when={didRepoSelectionChange}
            />
            <Form className="mt-4 d-flex" onSubmit={submit}>
                <LoaderButton
                    loading={fetchingRepos === 'loading'}
                    className="test-goto-add-external-service-page mr-2"
                    alwaysShowLabel={true}
                    type="submit"
                    label={fetchingRepos ? 'Saving...' : 'Save'}
                    disabled={fetchingRepos === 'loading' || !didRepoSelectionChange()}
                    variant="primary"
                />

                <Button
                    className="test-goto-add-external-service-page"
                    to={`${routingPrefix}/repositories`}
                    variant="secondary"
                    as={Link}
                >
                    Cancel
                </Button>
            </Form>
        </UserSettingReposContainer>
    )
}
