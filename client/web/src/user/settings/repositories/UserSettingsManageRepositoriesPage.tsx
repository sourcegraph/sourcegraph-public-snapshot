import classNames from 'classnames'
import { isEqual } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { FormEvent, useCallback, useEffect, useState, useRef } from 'react'
import { RouteComponentProps } from 'react-router'
import { Subscription } from 'rxjs'

import { Form } from '@sourcegraph/branded/src/components/Form'
import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { asError, ErrorLike, isErrorLike } from '@sourcegraph/shared/src/util/errors'
import { useRedesignToggle } from '@sourcegraph/shared/src/util/useRedesignToggle'
import { Container, PageSelector } from '@sourcegraph/wildcard'

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
    UserRepositoriesResult,
    SiteAdminRepositoryFields,
} from '../../../graphql-operations'
import {
    listUserRepositories,
    queryUserPublicRepositories,
    setUserPublicRepositories,
} from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserExternalServicesOrRepositoriesUpdateProps } from '../../../util'
import { externalServiceUserModeFromTags } from '../cloud-ga'

import { CheckboxRepositoryNode } from './RepositoryNode'

interface authenticatedUser {
    id: string
    siteAdmin: boolean
    tags: string[]
}

interface Props extends RouteComponentProps, TelemetryProps, UserExternalServicesOrRepositoriesUpdateProps {
    authenticatedUser: authenticatedUser
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
    <div key={warning} className="alert alert-warning mt-3 mb-0" role="alert">
        <AlertCircleIcon className="redesign-d-none icon icon-inline" /> {warning}. {hint}{' '}
        {hint ? 'for more details' : null}
    </div>
)

const displayError = (error: ErrorLike, hint?: JSX.Element): JSX.Element => (
    <div key={error.message} className="alert alert-danger mt-3 mb-0" role="alert">
        <AlertCircleIcon className="redesign-d-none icon icon-inline" /> {error.message}. {hint}{' '}
        {hint ? 'for more details' : null}
    </div>
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
export const UserSettingsManageRepositoriesPage: React.FunctionComponent<Props> = ({
    history,
    authenticatedUser,
    routingPrefix,
    telemetryService,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    // if we should tweak UI messaging and copy
    const ALLOW_PRIVATE_CODE = externalServiceUserModeFromTags(authenticatedUser.tags) === 'all'

    // if 'sync all' radio button is enabled and users can sync all repos from code hosts
    const ALLOW_SYNC_ALL = authenticatedUser.tags.includes('AllowUserExternalServiceSyncAll')

    // set up state hooks
    const [isRedesignEnabled] = useRedesignToggle()
    const [repoState, setRepoState] = useState(initialRepoState)
    const [publicRepoState, setPublicRepoState] = useState(initialPublicRepoState)
    const [codeHosts, setCodeHosts] = useState(initialCodeHostState)
    const [onloadSelectedRepos, setOnloadSelectedRepos] = useState<string[]>([])
    const [selectionState, setSelectionState] = useState(initialSelectionState)
    const [currentPage, setPage] = useState(1)
    const [query, setQuery] = useState('')
    const [codeHostFilter, setCodeHostFilter] = useState('')
    const [filteredRepos, setFilteredRepos] = useState<Repo[]>([])
    const [fetchingRepos, setFetchingRepos] = useState<initialFetchingReposState>()
    const externalServiceSubscription = useRef<Subscription>()

    // since we're making many different GraphQL requests - track affiliate and
    // manually added public repo errors separately
    const [affiliateRepoProblems, setAffiliateRepoProblems] = useState<affiliateRepoProblemType>()
    const [otherPublicRepoError, setOtherPublicRepoError] = useState<undefined | ErrorLike>()

    const ExternalServiceProblemHint = <Link to={`${routingPrefix}/code-hosts`}>Check code host connections</Link>

    const toggleTextArea = useCallback(
        () => setPublicRepoState({ ...publicRepoState, enabled: !publicRepoState.enabled }),
        [publicRepoState]
    )

    const fetchExternalServices = useCallback(
        async (): Promise<ExternalServicesResult['externalServices']['nodes']> =>
            queryExternalServices({
                first: null,
                after: null,
                namespace: authenticatedUser.id,
            })
                .toPromise()
                .then(({ nodes }) => nodes),

        [authenticatedUser.id]
    )

    const fetchAffiliatedRepos = useCallback(
        async (): Promise<AffiliatedRepositoriesResult['affiliatedRepositories']['nodes']> =>
            listAffiliatedRepositories({
                user: authenticatedUser.id,
                codeHost: null,
                query: null,
            })
                .toPromise()
                .then(({ affiliatedRepositories: { nodes } }) => nodes),

        [authenticatedUser.id]
    )

    const fetchSelectedRepositories = useCallback(
        async (): Promise<NonNullable<UserRepositoriesResult['node']>['repositories']['nodes']> =>
            listUserRepositories({ id: authenticatedUser.id, first: 2000 })
                .toPromise()
                .then(({ nodes }) => nodes),
        [authenticatedUser.id]
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

        // if external services may return code hosts with errors or warnings -
        // we can't safely continue
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
                // skip this code hots
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

        if (codeHostProblems.length > 0) {
            setAffiliateRepoProblems(codeHostProblems)
        }

        const [affiliatedRepos, selectedRepos] = await Promise.all([
            fetchAffiliatedRepos(),
            fetchSelectedRepositories(),
        ])

        const selectedAffiliatedRepos = new Map<string, Repo>()

        const affiliatedReposWithMirrorInfo = affiliatedRepos.map(affiliatedRepo => {
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
         * 3. if no repos selected - empty state
         */

        const radioSelectOption =
            ALLOW_SYNC_ALL &&
            ((externalServices.length === codeHostsHaveSyncAllQuery.length &&
                codeHostsHaveSyncAllQuery.every(Boolean)) ||
                affiliatedReposWithMirrorInfo.length === selectedAffiliatedRepos.size)
                ? 'all'
                : selectedAffiliatedRepos.size > 0
                ? 'selected'
                : ''

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
    }, [fetchExternalServices, fetchAffiliatedRepos, fetchSelectedRepositories, ALLOW_SYNC_ALL])

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
        const result = await queryUserPublicRepositories(authenticatedUser.id).toPromise()

        if (!result) {
            setPublicRepoState({ ...initialPublicRepoState, loaded: true })
        } else {
            // public repos separated by a new line
            const publicRepos = result.map(({ name }) => name)

            // safe off initial selection state
            setOnloadSelectedRepos(previousValue => [...previousValue, ...publicRepos])

            setPublicRepoState({ repos: publicRepos.join('\n'), loaded: true, enabled: result.length > 0 })
        }
    }, [authenticatedUser.id])

    useEffect(() => {
        fetchAndSetPublicRepos().catch(error => setOtherPublicRepoError(asError(error)))
    }, [fetchAndSetPublicRepos])

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

        return !isEqual(currentlySelectedRepos.sort(), onloadSelectedRepos.sort())
    }, [onloadSelectedRepos, publicRepoState.enabled, publicRepoState.repos, selectionState.repos])

    // save changes and update code hosts
    const submit = useCallback(
        async (event: FormEvent<HTMLFormElement>): Promise<void> => {
            event.preventDefault()
            eventLogger.log('UserManageRepositoriesSave')

            let publicRepos = publicRepoState.repos.split('\n').filter((row): boolean => row !== '')
            if (!publicRepoState.enabled) {
                publicRepos = []
            }

            setFetchingRepos('loading')

            try {
                await setUserPublicRepositories(authenticatedUser.id, publicRepos).toPromise()
            } catch (error) {
                setOtherPublicRepoError(asError(error))
                setFetchingRepos(undefined)
                return
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
            publicRepoState.repos,
            publicRepoState.enabled,
            authenticatedUser.id,
            codeHosts.hosts,
            selectionState.radio,
            selectionState.repos,
            history,
            routingPrefix,
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

    const hasProblems = affiliateRepoProblems !== undefined
    // code hosts were loaded and some were configured
    const hasCodeHosts = codeHosts.loaded && codeHosts.hosts.length !== 0
    const noCodeHostsOrErrors = !hasCodeHosts || hasProblems
    const hasCodeHostsNoErrors = hasCodeHosts && !hasProblems

    const modeSelect: JSX.Element = (
        <Form className="mt-4">
            <label className="d-flex flex-row align-items-baseline">
                <input
                    type="radio"
                    value="all"
                    disabled={!ALLOW_SYNC_ALL || noCodeHostsOrErrors}
                    checked={selectionState.radio === 'all'}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p
                        className={classNames('mb-0', {
                            'user-settings-repos__text': ALLOW_SYNC_ALL,
                            'user-settings-repos__text-disabled': !ALLOW_SYNC_ALL,
                        })}
                    >
                        Sync all repositories {!ALLOW_SYNC_ALL && '(coming soon)'}
                    </p>
                    <p
                        className={classNames({
                            'user-settings-repos__text': ALLOW_SYNC_ALL,
                            'user-settings-repos__text-disabled': !ALLOW_SYNC_ALL,
                        })}
                    >
                        Will sync all current and future public and private repositories
                    </p>
                </div>
            </label>
            <label className="d-flex flex-row align-items-baseline mb-0">
                <input
                    type="radio"
                    value="selected"
                    checked={selectionState.radio === 'selected'}
                    disabled={noCodeHostsOrErrors}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p
                        className={classNames({
                            'user-settings-repos__text-disabled': noCodeHostsOrErrors,
                            'mb-0': true,
                        })}
                    >
                        Sync selected {!ALLOW_PRIVATE_CODE && 'public'} repositories
                    </p>
                </div>
            </label>
        </Form>
    )

    const preventSubmit = useCallback((event: React.FormEvent<HTMLFormElement>): void => event.preventDefault(), [])

    const filterControls: JSX.Element = (
        <Form onSubmit={preventSubmit} className="w-100 d-inline-flex justify-content-between flex-row mt-3">
            <div className="d-inline-flex flex-row mr-3 align-items-baseline">
                <p className="text-xl-center text-nowrap mr-2">Code Host:</p>
                <select
                    className="form-control"
                    name="code-host"
                    aria-label="select code host type"
                    onChange={event => setCodeHostFilter(event.target.value)}
                >
                    <option key="any" value="" label="Any" />
                    {codeHosts.hosts.map(value => (
                        <option key={value.id} value={value.id} label={value.displayName} />
                    ))}
                </select>
            </div>
            <input
                className="form-control user-settings-repos__filter-input"
                type="search"
                placeholder="Search repositories..."
                name="query"
                autoComplete="off"
                autoCorrect="off"
                autoCapitalize="off"
                spellCheck={false}
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

    const selectAll = (): void => {
        const newMap = new Map<string, Repo>()
        // if not all repos are selected, we should select all, otherwise empty the selection

        if (selectionState.repos.size !== filteredRepos.length) {
            for (const repo of filteredRepos) {
                newMap.set(getRepoServiceAndName(repo), repo)
            }
        }
        setSelectionState({
            repos: newMap,
            loaded: selectionState.loaded,
            radio: selectionState.radio,
        })
    }

    const rows: JSX.Element = (
        <tbody>
            <tr className="align-items-baseline d-flex" key="header">
                <td className="user-settings-repos__repositorynode p-2 w-100 d-flex align-items-center border-top-0 border-bottom">
                    <input
                        id="select-all-repos"
                        className="mr-3"
                        type="checkbox"
                        checked={selectionState.repos.size !== 0 && selectionState.repos.size === filteredRepos.length}
                        onChange={selectAll}
                    />
                    <label
                        htmlFor="select-all-repos"
                        className={classNames({
                            'text-muted': selectionState.repos.size === 0,
                            'text-body': selectionState.repos.size !== 0,
                            'mb-0': true,
                        })}
                    >
                        {(selectionState.repos.size > 0 && (
                            <small>{`${selectionState.repos.size} ${
                                selectionState.repos.size === 1 ? 'repository' : 'repositories'
                            } selected`}</small>
                        )) || <small>Select all</small>}
                    </label>
                </td>
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
            <div className="mt-2 row">
                <div className="user-settings-repos__shimmer-circle mr-2" />
                <div className="user-settings-repos__shimmer mb-1 p-2 border-top-0 col-sm-2" />
            </div>
            <div className="mt-1 ml-2 row">
                <div className="user-settings-repos__shimmer mb-3 p-2 ml-1 border-top-0 col-sm-6" />
            </div>
            <div className="mt-2 row">
                <div className="user-settings-repos__shimmer-circle mr-2" />
                <div className="user-settings-repos__shimmer p-2 mb-1 border-top-0 col-sm-3" />
            </div>
        </div>
    )

    return (
        <div className="user-settings-repos">
            <PageTitle title="Manage Repositories" />
            <h2 className="mb-2">Manage Repositories</h2>
            <p className="text-muted">
                Choose repositories to sync with Sourcegraph to search code you care about all in one place
            </p>
            <Container>
                <ul className="list-group">
                    <li className="list-group-item user-settings-repos__container" key="from-code-hosts">
                        <div className={classNames(!isRedesignEnabled && 'p-4')}>
                            <h3>Your repositories</h3>
                            <p className="text-muted">
                                Repositories you own or collaborate on from your{' '}
                                <Link to={`${routingPrefix}/code-hosts`}>connected code hosts</Link>
                            </p>
                            {!ALLOW_PRIVATE_CODE && hasCodeHosts && (
                                <div className="alert alert-primary">
                                    Coming soon: search private repositories with Sourcegraph Cloud.{' '}
                                    <Link
                                        to="https://share.hsforms.com/1copeCYh-R8uVYGCpq3s4nw1n7ku"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    >
                                        Get updated when this feature launches
                                    </Link>
                                </div>
                            )}
                            {codeHosts.loaded && codeHosts.hosts.length === 0 && (
                                <div className="alert alert-warning mb-2">
                                    <Link to={`${routingPrefix}/code-hosts`}>Connect with a code host</Link> to add your
                                    own repositories to Sourcegraph.
                                </div>
                            )}
                            {displayAffiliateRepoProblems(affiliateRepoProblems, ExternalServiceProblemHint)}

                            {/* display radio buttons shimmer only when user has code hosts */}
                            {hasCodeHostsNoErrors && !selectionState.loaded && modeSelectShimmer}

                            {/* display type of repo sync radio buttons */}
                            {hasCodeHostsNoErrors && selectionState.loaded && modeSelect}

                            {
                                // if we're in 'selected' mode, show a list of all the repos on the code hosts to select from
                                hasCodeHostsNoErrors && selectionState.radio === 'selected' && (
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
                    </li>
                    {window.context.sourcegraphDotComMode && (
                        <li className="list-group-item user-settings-repos__container" key="add-textarea">
                            <div className={classNames(!isRedesignEnabled && 'p-4')}>
                                <h3>Other public repositories</h3>
                                <p className="text-muted">Public repositories on GitHub and GitLab</p>
                                <input
                                    id="add-public-repos"
                                    className="mr-2 mt-2"
                                    type="checkbox"
                                    onChange={toggleTextArea}
                                    checked={publicRepoState.enabled}
                                />
                                <label htmlFor="add-public-repos">Sync specific public repositories by URL</label>

                                {publicRepoState.enabled && (
                                    <div className="form-group ml-4 mt-3">
                                        <p className="mb-2">Repositories to sync</p>
                                        <textarea
                                            className="form-control"
                                            rows={5}
                                            value={publicRepoState.repos}
                                            onChange={handlePublicReposChanged}
                                        />
                                        <p className="text-muted mt-2">
                                            Specify with complete URLs. One repository per line.
                                        </p>
                                    </div>
                                )}
                            </div>
                        </li>
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
                    className="btn btn-primary test-goto-add-external-service-page mr-2"
                    alwaysShowLabel={true}
                    type="submit"
                    label={fetchingRepos ? 'Saving...' : 'Save'}
                    disabled={fetchingRepos === 'loading'}
                />

                <Link
                    className="btn btn-secondary test-goto-add-external-service-page"
                    to={`${routingPrefix}/repositories`}
                >
                    Cancel
                </Link>
            </Form>
        </div>
    )
}
