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
import { repeatUntil } from '@sourcegraph/shared/src/util/rxjs/repeatUntil'
import { PageSelector } from '@sourcegraph/wildcard'

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
} from '../../../graphql-operations'
import { queryUserPublicRepositories, setUserPublicRepositories } from '../../../site-admin/backend'
import { eventLogger } from '../../../tracking/eventLogger'
import { UserRepositoriesUpdateProps } from '../../../util'

import { AwayPrompt, ALLOW_NAVIGATION } from './AwayPrompt'
import { CheckboxRepositoryNode } from './RepositoryNode'

interface Props extends RouteComponentProps, TelemetryProps, UserRepositoriesUpdateProps {
    userID: string
    routingPrefix: string
}

interface Repo {
    name: string
    codeHost: Maybe<{ kind: ExternalServiceKind; id: string; displayName: string }>
    private: boolean
}

interface GitHubConfig {
    repos: string[]
    token: 'REDACTED'
    url: string
}
interface GitLabConfig {
    projectQuery: string[]
    projects: { name: string }[]
    token: 'REDACTED'
    url: string
}

const PER_PAGE = 25
const SIX_SECONDS = 6000
const EIGHT_SECONDS = 8000

// initial state constants
const emptyRepos: Repo[] = []
const initialRepoState = {
    repos: emptyRepos,
    loading: false,
    loaded: false,
}

const emptyHosts: ExternalServicesResult['externalServices']['nodes'] = []
const emptyRepoNames: string[] = []
const initialCodeHostState = {
    hosts: emptyHosts,
    loaded: false,
    configuredRepos: emptyRepoNames,
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

type initialFetchingReposState = undefined | 'loading' | 'slow' | 'slower'
type affiliateRepoProblemType = undefined | string | ErrorLike | ErrorLike[]

const isLoading = (status: initialFetchingReposState): boolean => {
    if (!status) {
        return false
    }

    return ['loading', 'slow', 'slower'].includes(status)
}

const displayWarning = (warning: string, hint?: JSX.Element): JSX.Element => (
    <div key={warning} className="alert alert-warning mt-3" role="alert">
        <AlertCircleIcon className="icon icon-inline" /> {warning}. {hint} {hint ? 'for more details' : null}
    </div>
)

const displayError = (error: ErrorLike, hint?: JSX.Element): JSX.Element => (
    <div key={error.message} className="alert alert-danger mt-3" role="alert">
        <AlertCircleIcon className="icon icon-inline" /> {error.message}. {hint} {hint ? 'for more details' : null}
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
    userID,
    routingPrefix,
    telemetryService,
    onUserRepositoriesUpdate,
}) => {
    useEffect(() => {
        telemetryService.logViewEvent('UserSettingsRepositories')
    }, [telemetryService])

    // set up state hooks
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

    const ExternalServiceProblemHint = (
        <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
            Check code host connections
        </Link>
    )

    const toggleTextArea = useCallback(
        () => setPublicRepoState({ ...publicRepoState, enabled: !publicRepoState.enabled }),
        [publicRepoState]
    )

    const fetchAndSetExternalServices = useCallback(async (): Promise<void> => {
        const result = await queryExternalServices({
            first: null,
            after: null,
            namespace: userID,
        }).toPromise()

        const selected: string[] = []
        // if external services may return code hosts with errors or warnings -
        // we can't safely continue
        const codeHostProblems = []

        for (const host of result.nodes) {
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
                    if (gitLabCfg.projects !== undefined) {
                        gitLabCfg.projects.map(project => {
                            selected.push(project.name)
                        })
                    }
                    break
                }

                case ExternalServiceKind.GITHUB: {
                    const gitHubCfg = cfg as GitHubConfig
                    if (gitHubCfg.repos !== undefined) {
                        selected.push(...gitHubCfg.repos)
                    }
                    break
                }
            }
        }

        if (codeHostProblems.length > 0) {
            setAffiliateRepoProblems(codeHostProblems)
        }

        setCodeHosts({
            loaded: true,
            hosts: result.nodes,
            configuredRepos: selected,
        })

        if (selected.length !== 0) {
            // if user's code hosts have repos - collapse affiliated repos
            // section
            setSelectionState(previousSelectionState => ({
                ...previousSelectionState,
                radio: 'selected',
            }))
        }
    }, [userID, setCodeHosts])

    // fetch public repos for the "other public repositories" textarea
    const fetchAndSetPublicRepos = useCallback(async (): Promise<void> => {
        const result = await queryUserPublicRepositories(userID).toPromise()

        if (!result) {
            setPublicRepoState({ ...initialPublicRepoState, loaded: true })
        } else {
            // public repos separated by a new line
            const publicRepos = result.map(({ name }) => name)

            // safe off initial selection state
            setOnloadSelectedRepos(previousValue => [...previousValue, ...publicRepos])

            setPublicRepoState({ repos: publicRepos.join('\n'), loaded: true, enabled: result.length > 0 })
        }
    }, [userID])

    useEffect(() => {
        fetchAndSetExternalServices().catch(error => {
            setAffiliateRepoProblems(asError(error))
        })
    }, [fetchAndSetExternalServices])

    const fetchAffiliatedRepos = useCallback(
        async (): Promise<AffiliatedRepositoriesResult> =>
            listAffiliatedRepositories({
                user: userID,
                codeHost: null,
                query: null,
            }).toPromise(),

        [userID]
    )

    useEffect(() => {
        // once we've loaded code hosts and the 'selected' panel is visible
        // load repos and set the loading state
        if (selectionState.radio === 'selected' && codeHosts.loaded) {
            // trigger shimmer effect
            setRepoState(previousRepoState => ({
                ...previousRepoState,
                loading: true,
            }))

            fetchAffiliatedRepos()
                .then(result => {
                    const { nodes: affiliatedRepos } = result.affiliatedRepositories

                    const selectedRepos = new Map<string, Repo>()

                    // create a map of user selected affiliated repos
                    for (const repoName of codeHosts.configuredRepos) {
                        const affiliatedRepo = affiliatedRepos.find(repo => repo.name === repoName)
                        if (affiliatedRepo) {
                            selectedRepos.set(repoName, affiliatedRepo)
                        }
                    }

                    // sort affiliated repos with already selected repos at the top
                    affiliatedRepos.sort((repoA, repoB): number => {
                        const isRepoASelected = selectedRepos.has(repoA.name)
                        const isRepoBSelected = selectedRepos.has(repoB.name)

                        if (!isRepoASelected && isRepoBSelected) {
                            return 1
                        }

                        if (isRepoASelected && !isRepoBSelected) {
                            return -1
                        }

                        return 0
                    })

                    setRepoState(previousRepoState => ({
                        ...previousRepoState,
                        repos: affiliatedRepos,
                        loaded: true,
                    }))

                    // safe off initial selection state
                    setOnloadSelectedRepos(previousValue => [...previousValue, ...selectedRepos.keys()])

                    setSelectionState({
                        repos: selectedRepos,
                        radio: selectionState.radio,
                        loaded: true,
                    })
                })
                .catch(error => {
                    setAffiliateRepoProblems(asError(error))
                    setRepoState({
                        repos: emptyRepos,
                        loading: false,
                        loaded: true,
                    })
                })
        }
    }, [selectionState.radio, codeHosts.loaded, codeHosts.configuredRepos, fetchAffiliatedRepos])

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
                await setUserPublicRepositories(userID, publicRepos).toPromise()
            } catch (error) {
                setOtherPublicRepoError(asError(error))
                setFetchingRepos(undefined)
                return
            }

            if (!selectionState.radio) {
                // location state is used here to prevent AwayPrompt from blocking
                return history.push(routingPrefix + '/repositories', ALLOW_NAVIGATION)
            }

            const syncTimes = new Map<string, string | null>()
            const codeHostRepoPromises = []

            for (const host of codeHosts.hosts) {
                const repos: string[] = []
                syncTimes.set(host.id, host.lastSyncAt)
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

            const started = Date.now()
            externalServiceSubscription.current = queryExternalServices({
                first: null,
                after: null,
                namespace: userID,
            })
                .pipe(
                    repeatUntil(
                        result => {
                            // if the background job takes too long we should update the button
                            // text to indicate we're still working on it.

                            const now = Date.now()
                            const timeDiff = now - started

                            // setting the same state multiple times won't cause
                            // re-renders in Function components
                            if (timeDiff >= SIX_SECONDS + EIGHT_SECONDS) {
                                setFetchingRepos('slower')
                            } else if (timeDiff >= SIX_SECONDS) {
                                setFetchingRepos('slow')
                            }

                            // if the lastSyncAt has changed for all hosts then we're done
                            if (
                                result.nodes.every(
                                    codeHost =>
                                        codeHost.lastSyncAt && codeHost.lastSyncAt !== syncTimes.get(codeHost.id)
                                )
                            ) {
                                const repoCount = result.nodes.reduce((sum, codeHost) => sum + codeHost.repoCount, 0)
                                onUserRepositoriesUpdate(repoCount)

                                // push the user back to the repo list page
                                // location state is used here to prevent AwayPrompt from blocking
                                history.push(routingPrefix + '/repositories', ALLOW_NAVIGATION)

                                // cancel the repeatUntil
                                return true
                            }
                            // keep repeating
                            return false
                        },
                        { delay: 2000 }
                    )
                )
                .subscribe(
                    () => {},
                    error => setAffiliateRepoProblems(asError(error)),
                    () => {
                        externalServiceSubscription.current?.unsubscribe()
                    }
                )
        },
        [
            publicRepoState.repos,
            publicRepoState.enabled,
            userID,
            codeHosts.hosts,
            selectionState.radio,
            selectionState.repos,
            onUserRepositoriesUpdate,
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

    const modeSelect: JSX.Element = (
        <Form className="mt-4">
            <label className="d-flex flex-row align-items-baseline">
                <input
                    type="radio"
                    value="all"
                    disabled={true}
                    checked={selectionState.radio === 'all'}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p
                        className="mb-0 user-settings-repos__text-disabled
"
                    >
                        Sync all repositories (coming soon)
                    </p>
                    <p
                        className="user-settings-repos__text-disabled
"
                    >
                        Will sync all current and future public and private repositories
                    </p>
                </div>
            </label>
            <label className="d-flex flex-row align-items-baseline">
                {/* TODO: @artem add this functionality after pe-GA */}
                <input
                    type="radio"
                    value="all"
                    disabled={true}
                    checked={selectionState.radio === 'org'}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p
                        className="mb-0 user-settings-repos__text-disabled
"
                    >
                        Sync all repositories from selected organizations or users (coming soon)
                    </p>
                    <p
                        className="user-settings-repos__text-disabled
"
                    >
                        Will sync all current and future public and private repositories
                    </p>
                </div>
            </label>
            <label className="d-flex flex-row align-items-baseline">
                <input
                    type="radio"
                    value="selected"
                    checked={selectionState.radio === 'selected'}
                    disabled={affiliateRepoProblems !== undefined || codeHosts.hosts.length === 0}
                    onChange={handleRadioSelect}
                />
                <div className="d-flex flex-column ml-2">
                    <p
                        className={classNames({
                            'user-settings-repos__text-disabled':
                                affiliateRepoProblems !== undefined || codeHosts.hosts.length === 0,
                            'mb-0': true,
                        })}
                    >
                        Sync selected public repositories
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
            const newMap = new Map(selectionState.repos)
            if (newMap.has(repo.name)) {
                newMap.delete(repo.name)
            } else {
                newMap.set(repo.name, repo)
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
                newMap.set(repo.name, repo)
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
                return (
                    <CheckboxRepositoryNode
                        name={repo.name}
                        key={repo.name}
                        onClick={onRepoClicked(repo)}
                        checked={selectionState.repos.has(repo.name)}
                        serviceType={repo.codeHost?.kind || ''}
                        isPrivate={repo.private}
                    />
                )
            })}
        </tbody>
    )

    const handlePublicReposChanged = (event: React.ChangeEvent<HTMLTextAreaElement>): void => {
        setPublicRepoState({ ...publicRepoState, repos: event.target.value })
    }

    const loadingAnimation: JSX.Element = (
        <tbody className="container">
            <tr className="mt-2 align-items-baseline row">
                <td className="user-settings-repos__shimmer p-3 border-top-0 col-sm-9" />
            </tr>
            <tr className="mt-2 align-items-baseline row">
                <td className="user-settings-repos__shimmer p-3 border-top-0 col-sm-4" />
            </tr>
            <tr className="mt-2 align-items-baseline row">
                <td className="user-settings-repos__shimmer p-3 border-top-0 col-sm-7" />
            </tr>
        </tbody>
    )
    return (
        <div className="user-settings-repos">
            <PageTitle title="Manage Repositories" />
            <h2 className="mb-2">Manage Repositories</h2>
            <p className="text-muted">
                Choose repositories to sync with Sourcegraph to search code you care about all in one place
            </p>
            <ul className="list-group">
                <li className="list-group-item p-0 user-settings-repos__container" key="from-code-hosts">
                    <div className="p-4">
                        <h3>Your repositories</h3>
                        <p className="text-muted">
                            Repositories you own or collaborate on from your{' '}
                            <Link className="text-primary" to={`${routingPrefix}/code-hosts`}>
                                connected code hosts
                            </Link>
                        </p>
                        {codeHosts.loaded && codeHosts.hosts.length !== 0 && (
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
                                <Link to={`${routingPrefix}/code-hosts`}>Connect with a code host</Link> to add your own
                                repositories to Sourcegraph.
                            </div>
                        )}
                        {displayAffiliateRepoProblems(affiliateRepoProblems, ExternalServiceProblemHint)}
                        {codeHosts.loaded &&
                            codeHosts.hosts.length !== 0 &&
                            // display radio button for 'all' or 'selected' repos
                            modeSelect}
                        {
                            // if we're in 'selected' mode, show a list of all the repos on the code hosts to select from
                            selectionState.radio === 'selected' && (
                                <div className="ml-4">
                                    {filterControls}
                                    <table role="grid" className="table">
                                        {
                                            // if we're selecting repos, and the repos are still loading, display the loading animation
                                            selectionState.radio === 'selected' &&
                                                repoState.loading &&
                                                !repoState.loaded &&
                                                loadingAnimation
                                        }
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
                    <li className="list-group-item p-0 user-settings-repos__container" key="add-textarea">
                        <div className="p-4">
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
            {isErrorLike(otherPublicRepoError) && displayError(otherPublicRepoError)}
            <AwayPrompt
                header="Discard unsaved changes?"
                message="Currently synced repositories will be unchanged"
                button_ok_text="Discard"
                when={didRepoSelectionChange}
            />
            <Form className="mt-4 d-flex" onSubmit={submit}>
                <LoaderButton
                    loading={isLoading(fetchingRepos)}
                    className="btn btn-primary test-goto-add-external-service-page mr-2"
                    alwaysShowLabel={true}
                    type="submit"
                    label={
                        (!fetchingRepos && 'Save') ||
                        (fetchingRepos === 'loading' && 'Saving...') ||
                        (fetchingRepos === 'slow' && 'Still saving...') ||
                        'Any time now...'
                    }
                    disabled={isLoading(fetchingRepos)}
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
