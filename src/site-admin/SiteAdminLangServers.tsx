import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { sortBy } from 'lodash'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import BugIcon from 'mdi-react/BugIcon'
import DownloadIcon from 'mdi-react/DownloadIcon'
import GithubCircleIcon from 'mdi-react/GithubCircleIcon'
import RefreshIcon from 'mdi-react/RefreshIcon'
import WebIcon from 'mdi-react/WebIcon'
import * as React from 'react'
import { concat, merge, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, delay, filter, map, repeatWhen, switchMap, tap } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { eventLogger } from '../tracking/eventLogger'
import { disableLangServer, enableLangServer, fetchLangServers, restartLangServer, updateLangServer } from './backend'

interface Props {}

type LanguageState = 'updating' | 'restarting' | 'enabling' | 'disabling'

interface State {
    langServers: GQL.ILangServer[]
    loading: boolean
    error?: Error

    /**
     * Maps languages to an error occurring about that language. e.g. if
     * updating a specific language fails, an error will be present here for
     * that language.
     */
    errorsBylanguage: Map<string, Error>

    /**
     * Maps languages to their current pending state. e.g. if the Restart
     * button is clicked, but the GraphQL mutation has not returned yet, this
     * state will indicate that the language is 'restarting'.
     */
    pendingStateByLanguage: Map<string, LanguageState>
}

/**
 * Component to show the status of language servers.
 */
export class SiteAdminLangServers extends React.PureComponent<Props, State> {
    public state: State = {
        langServers: [],
        loading: false,
        errorsBylanguage: new Map<string, Error>(),
        pendingStateByLanguage: new Map<string, LanguageState>(),
    }

    private subscriptions = new Subscription()
    private refreshLangServers = new Subject<void>()

    private updateButtonClicks = new Subject<GQL.ILangServer>()
    private restartButtonClicks = new Subject<GQL.ILangServer>()
    private disableButtonClicks = new Subject<GQL.ILangServer>()
    private enableButtonClicks = new Subject<GQL.ILangServer>()

    private EXPERIMENTAL_LANGUAGE_SERVER_WARNING =
        'This language server is experimental and under active development. ' +
        'Be aware that it may run arbitrary code through the package manager ' +
        'during dependency installation for each repository. Some extensions ' +
        'to the language server protocol such as cross-repository code ' +
        'intelligence might not be available. Are you sure you want to ' +
        'enable it?'

    public componentDidMount(): void {
        this.subscriptions.add(
            merge(
                this.updateButtonClicks.pipe(
                    tap(langServer => this.logClick('LangServerUpdateClicked', langServer)),
                    map(langServer => ({
                        langServer,
                        mutation: updateLangServer,
                        errorEventLabel: 'LangServersUpdateFailed',
                        stateKey: 'updating' as 'updating',
                    }))
                ),
                this.restartButtonClicks.pipe(
                    tap(langServer => this.logClick('LangServerRestartClicked', langServer)),
                    map(langServer => ({
                        langServer,
                        mutation: restartLangServer,
                        errorEventLabel: 'LangServersRestartFailed',
                        stateKey: 'restarting' as 'restarting',
                    }))
                ),
                this.disableButtonClicks.pipe(
                    tap(langServer => this.logClick('LangServerDisableClicked', langServer)),
                    map(langServer => ({
                        langServer,
                        mutation: disableLangServer,
                        errorEventLabel: 'LangServersDisableFailed',
                        stateKey: 'disabling' as 'disabling',
                    }))
                ),
                this.enableButtonClicks.pipe(
                    tap(langServer => this.logClick('LangServerEnableClicked', langServer)),
                    filter(
                        langServer =>
                            langServer.experimental ? window.confirm(this.EXPERIMENTAL_LANGUAGE_SERVER_WARNING) : true
                    ),
                    map(langServer => ({
                        langServer,
                        mutation: enableLangServer,
                        errorEventLabel: 'LangServersEnableFailed',
                        stateKey: 'enabling' as 'enabling',
                    }))
                )
            )
                .pipe(
                    tap(({ langServer, stateKey }) => {
                        this.setState(prevState => {
                            const newErrorsByLanguage = new Map(this.state.errorsBylanguage)
                            newErrorsByLanguage.delete(langServer.language)
                            return {
                                pendingStateByLanguage: new Map(this.state.pendingStateByLanguage).set(
                                    langServer.language,
                                    stateKey
                                ),
                                errorsBylanguage: newErrorsByLanguage,
                            }
                        })
                    }),
                    concatMap(({ langServer, mutation, errorEventLabel }) =>
                        mutation(langServer.language).pipe(
                            map(() => ({
                                langServer,
                                newState: (prevState: State): Partial<State> => ({}),
                            })),
                            catchError(error => {
                                eventLogger.log(errorEventLabel, {
                                    lang_server: { error_message: error.message },
                                })
                                console.error(error)
                                return [
                                    {
                                        langServer,
                                        newState: (prevState: State): Partial<State> => ({
                                            errorsBylanguage: new Map(prevState.errorsBylanguage).set(
                                                langServer.language,
                                                error
                                            ),
                                        }),
                                    },
                                ]
                            })
                        )
                    )
                )
                .subscribe(
                    ({ langServer, newState }) => {
                        this.refreshLangServers.next()
                        const newPendingStateByLanguage = this.state.pendingStateByLanguage
                        newPendingStateByLanguage.delete(langServer.language)
                        this.setState(prevState => ({
                            ...prevState,
                            ...newState(prevState),
                            pendingStateByLanguage: newPendingStateByLanguage,
                        }))
                    },
                    err => console.error(err)
                )
        )

        this.subscriptions.add(
            // Trigger poll sequences initially and on every button click
            concat([undefined], this.refreshLangServers)
                .pipe(
                    // When we forced a refresh, cancel the previous
                    // in-flight request and start a new poll sequence
                    switchMap(() =>
                        fetchLangServers().pipe(
                            // Map to state update
                            map(langServers => ({ langServers, error: undefined, loading: false })),
                            catchError(error => {
                                eventLogger.log('LangServersFetchFailed', {
                                    langServers: { error_message: error.message },
                                })
                                console.error(error)
                                return [{ langServers: [], error, loading: false }]
                            }),
                            // After the request completed, resubscribe after 5s to trigger another request
                            repeatWhen(delay(2500))
                        )
                    )
                )
                .subscribe(newState => {
                    this.setState(newState)
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className="site-admin-lang-servers mb-3">
                <div className="site-admin-lang-servers__header">
                    <div className="site-admin-lang-servers__header-icon">
                        <WebIcon className="icon-inline" />
                    </div>
                    <h5 className="site-admin-lang-servers__header-title">Language servers</h5>
                </div>
                {!this.state.error &&
                    this.state.langServers.length === 0 && (
                        <LoadingSpinner className="site-admin-lang-servers__loading-icon" />
                    )}
                {this.state.error && (
                    <div className="site-admin-lang-servers__error">
                        <AlertCircleIcon className="icon-inline" />
                        <span className="site-admin-lang-servers__error-text">Error: {this.state.error.message}</span>
                    </div>
                )}
                {// Sort the language servers in a stable fashion such that
                // experimental ones are at the bottom of the list.
                sortBy(this.state.langServers, langServer => (langServer.experimental ? 1 : 0)).map((langServer, i) => (
                    <div className="site-admin-lang-servers__list-item" key={i}>
                        <div className="site-admin-lang-servers__left-area">
                            <div className="site-admin-lang-servers__language">
                                <div className="site-admin-lang-servers__language-name">{langServer.displayName}</div>
                                {langServer.experimental && (
                                    <span
                                        className="badge badge-warning"
                                        data-tooltip="This language server is experimental. Beware it may run arbitrary code and might have limited functionality."
                                    >
                                        experimental
                                    </span>
                                )}
                                {langServer.custom && (
                                    <span
                                        className="site-admin-lang-servers__language-custom"
                                        data-tooltip="This language server is custom / does not come built in with Sourcegraph. It was added via the site configuration."
                                    >
                                        (custom)
                                    </span>
                                )}
                            </div>
                            {this.renderRepo(langServer)}
                        </div>
                        <div>
                            {this.renderActions(langServer)}
                            {this.renderStatus(langServer)}
                        </div>
                    </div>
                ))}
            </div>
        )
    }

    private renderStatus(langServer: GQL.ILangServer): JSX.Element | null {
        // If any action is currently pending, then disregard the langserver
        // state we have from the backend and just display the pending
        // indicator.
        if (this.state.pendingStateByLanguage.has(langServer.language)) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--pending">
                    <LoadingSpinner className="icon-inline" />
                </span>
            )
        }

        if (langServer.state === 'LANG_SERVER_STATE_NONE') {
            return null
        }
        if (langServer.state === 'LANG_SERVER_STATE_DISABLED') {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--disabled">
                    ○ Disabled
                </span>
            )
        }

        // TODO(chris): Remove this when more robust health checking is
        // available (e.g. TCP checks).
        const couldDetermineHealth =
            langServer.pending ||
            langServer.canEnable ||
            langServer.canDisable ||
            langServer.canRestart ||
            langServer.canUpdate ||
            langServer.healthy

        // In cluster deployments (and in dev mode without DEBUG_MANAGE_DOCKER=t) the
        // docker socket is unavailable, which prevents the backend from running
        // `docker inspect` to determine if the language server is healthy.
        // Custom language servers are not managed by our infrastructure, so we
        // have no insight into their state other than the associated TCP
        // connection or stdio handles to the process. This could be improved in
        // the future by checking that the TCP connection is still open or that
        // the process is still running.
        if (langServer.isClusterDeployment || langServer.custom || !couldDetermineHealth) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--running">
                    ● Enabled
                </span>
            )
        }

        // Code past here uses fields that are not present for clusters
        if (langServer.pending) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--pending">
                    <LoadingSpinner className="icon-inline" />
                </span>
            )
        }
        if (langServer.healthy) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--running">
                    ● Running
                </span>
            )
        }
        return (
            <span className="site-admin-lang-servers__status site-admin-lang-servers__status--unhealthy">
                ● Unhealthy
            </span>
        )
    }

    private renderActions = (langServer: GQL.ILangServer) => {
        // Disable the action buttons while any of these conditions hold:
        //
        // - The state is expected to change (e.g. after clicking one of the
        //   action buttons)
        // - The language server is currently being downloaded
        const disabled = this.state.pendingStateByLanguage.has(langServer.language) || langServer.downloading
        const updating =
            this.state.pendingStateByLanguage.has(langServer.language) &&
            this.state.pendingStateByLanguage.get(langServer.language) === 'updating'
        return (
            <div className="site-admin-lang-servers__actions btn-group" role="group">
                {updating && (
                    <span className="site-admin-lang-servers__actions-updating">Pulling latest Docker image…</span>
                )}
                {langServer.canUpdate && (
                    <button
                        disabled={disabled}
                        type="button"
                        className="site-admin-lang-servers__actions-update btn btn-secondary"
                        data-tooltip={!disabled ? 'Update language server' : undefined}
                        // tslint:disable-next-line:jsx-no-lambda
                        onClick={() => this.updateButtonClicks.next(langServer)}
                    >
                        <DownloadIcon className="icon-inline" />
                    </button>
                )}
                {langServer.state !== 'LANG_SERVER_STATE_DISABLED' &&
                    langServer.canRestart && (
                        <button
                            disabled={disabled}
                            type="button"
                            className="site-admin-lang-servers__actions-restart btn btn-secondary"
                            data-tooltip={!disabled ? 'Restart language server' : undefined}
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.restartButtonClicks.next(langServer)}
                        >
                            <RefreshIcon className="icon-inline" />
                        </button>
                    )}
                {langServer.state === 'LANG_SERVER_STATE_ENABLED' &&
                    langServer.canDisable && (
                        <button
                            disabled={disabled}
                            type="button"
                            className="site-admin-lang-servers__actions-enable-disable btn btn-secondary"
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.disableButtonClicks.next(langServer)}
                        >
                            Disable
                        </button>
                    )}
                {(langServer.state === 'LANG_SERVER_STATE_DISABLED' || langServer.state === 'LANG_SERVER_STATE_NONE') &&
                    langServer.canEnable && (
                        <button
                            disabled={disabled}
                            type="button"
                            className="btn btn-secondary site-admin-lang-servers__actions-enable-disable"
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.enableButtonClicks.next(langServer)}
                        >
                            Enable
                        </button>
                    )}
            </div>
        )
    }

    private renderRepo = (langServer: GQL.ILangServer) => {
        if (!langServer.homepageURL && !langServer.docsURL && !langServer.issuesURL) {
            return null
        }
        return (
            <small>
                <div className="site-admin-lang-servers__repo">
                    {langServer.homepageURL &&
                        langServer.homepageURL.startsWith('https://github.com') && (
                            <>
                                <a
                                    className="site-admin-lang-servers__repo-link"
                                    href={langServer.homepageURL}
                                    target="_blank"
                                    onClick={this.generateClickHandler('LangServerHomepageClicked', langServer)}
                                >
                                    <GithubCircleIcon className="icon-inline" />{' '}
                                    {langServer.homepageURL.substr('https://github.com/'.length)}
                                </a>
                            </>
                        )}
                    {langServer.homepageURL &&
                        !langServer.homepageURL.startsWith('https://github.com') && (
                            <a
                                className="site-admin-lang-servers__repo-link"
                                href={langServer.homepageURL}
                                target="_blank"
                                onClick={this.generateClickHandler('LangServerHomepageClicked', langServer)}
                            >
                                {langServer.homepageURL}
                            </a>
                        )}
                    {langServer.issuesURL && (
                        <a
                            className="site-admin-lang-servers__repo-link"
                            href={langServer.issuesURL}
                            target="_blank"
                            data-tooltip="View issues"
                            onClick={this.generateClickHandler('LangServerIssuesClicked', langServer)}
                        >
                            <BugIcon className="icon-inline" />
                        </a>
                    )}
                </div>
            </small>
        )
    }

    private logClick(eventLabel: string, langServer: GQL.ILangServer): void {
        eventLogger.log(eventLabel, {
            // 🚨 PRIVACY: never provide any private data in { lang_server: { ... } }.
            lang_server: {
                language: langServer.language,
            },
        })
    }

    private generateClickHandler(eventLabel: string, langServer: GQL.ILangServer): () => void {
        return () => this.logClick(eventLabel, langServer)
    }
}
