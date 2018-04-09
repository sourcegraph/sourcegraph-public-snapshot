import BugIcon from '@sourcegraph/icons/lib/Bug'
import DownloadSimpleIcon from '@sourcegraph/icons/lib/DownloadSimple'
import ErrorIcon from '@sourcegraph/icons/lib/Error'
import FileDocumentBoxIcon from '@sourcegraph/icons/lib/FileDocumentBox'
import GitHubIcon from '@sourcegraph/icons/lib/GitHub'
import GlobeIcon from '@sourcegraph/icons/lib/Globe'
import LoaderIcon from '@sourcegraph/icons/lib/Loader'
import RefreshIcon from '@sourcegraph/icons/lib/Refresh'
import * as React from 'react'
import { interval } from 'rxjs/observable/interval'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { concatMap } from 'rxjs/operators/concatMap'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../tracking/eventLogger'
import { disableLangServer, enableLangServer, fetchLangServers, restartLangServer, updateLangServer } from './backend'

interface Props {}

type LanguageState = 'updating' | 'restarting' | 'enabling' | 'disabling'

interface State {
    langServers: GQL.ILangServer[]
    loading: boolean
    error?: Error

    /**
     * Maps languages to an error occuring about that language. e.g. if
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

    public componentDidMount(): void {
        this.subscriptions.add(
            merge(
                this.updateButtonClicks.pipe(
                    map(langServer => ({
                        langServer,
                        mutation: updateLangServer,
                        errorEventLabel: 'LangServersUpdateFailed',
                        stateKey: 'updating' as 'updating',
                    }))
                ),
                this.restartButtonClicks.pipe(
                    map(langServer => ({
                        langServer,
                        mutation: restartLangServer,
                        errorEventLabel: 'LangServersRestartFailed',
                        stateKey: 'restarting' as 'restarting',
                    }))
                ),
                this.disableButtonClicks.pipe(
                    map(langServer => ({
                        langServer,
                        mutation: disableLangServer,
                        errorEventLabel: 'LangServersDisableFailed',
                        stateKey: 'disabling' as 'disabling',
                    }))
                ),
                this.enableButtonClicks.pipe(
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
            merge(this.refreshLangServers, interval(2500))
                .pipe(
                    startWith<void | number>(0),
                    switchMap(() =>
                        fetchLangServers().pipe(
                            map(langServers => ({
                                langServers,
                                error: undefined,
                                loading: false,
                            })),
                            catchError(error => {
                                eventLogger.log('LangServersFetchFailed', {
                                    langServers: { error_message: error.message },
                                })
                                console.error(error)
                                return [{ langServers: [], error, loading: false }]
                            })
                        )
                    )
                )
                .subscribe(
                    newState => {
                        this.setState(newState)
                    },
                    err => console.error(err)
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <ul className="site-admin-lang-servers">
                <div className="site-admin-lang-servers__header">
                    <div className="site-admin-lang-servers__header-icon">
                        <GlobeIcon className="icon-inline" />
                    </div>
                    <h5 className="site-admin-lang-servers__header-title">Language Servers</h5>
                </div>
                {!this.state.error &&
                    this.state.langServers.length === 0 && (
                        <LoaderIcon className="site-admin-lang-servers__loading-icon" />
                    )}
                {this.state.error && (
                    <div className="site-admin-lang-servers__error">
                        <ErrorIcon className="icon-inline" />
                        <span className="site-admin-lang-servers__error-text">Error: {this.state.error.message}</span>
                    </div>
                )}
                {this.state.langServers.map((langServer, i) => (
                    <div className="site-admin-lang-servers__list-item" key={i}>
                        <div className="site-admin-lang-servers__left-area">
                            <div className="site-admin-lang-servers__language">
                                {langServer.displayName}
                                {langServer.custom && (
                                    <span
                                        className="site-admin-lang-servers__language-custom"
                                        data-tooltip="This language server is custom / does not come built in with Sourcegraph. It was added via the site configuration."
                                    >
                                        (custom)
                                    </span>
                                )}
                                {this.renderStatus(langServer)}
                            </div>
                            {this.renderRepo(langServer)}
                        </div>
                        {this.renderActions(langServer)}
                    </div>
                ))}
            </ul>
        )
    }

    private renderStatus(langServer: GQL.ILangServer): JSX.Element | null {
        // If any action is currently pending, then disregard the langserver
        // state we have from the backend and just display the pending
        // indicator.
        if (this.state.pendingStateByLanguage.has(langServer.language)) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--pending">
                    <LoaderIcon className="icon-inline" />
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

        // If we're running in data center mode OR the language server is
        // custom, then all we know at this point is that it is enabled.
        if (langServer.dataCenter || langServer.custom) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--running">
                    ● Enabled
                </span>
            )
        }

        // Code past here uses fields that are not present in Data Center mode.
        if (langServer.pending) {
            return (
                <span className="site-admin-lang-servers__status site-admin-lang-servers__status--pending">
                    <LoaderIcon className="icon-inline" />
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
        const disabled = this.state.pendingStateByLanguage.has(langServer.language)
        const updating =
            this.state.pendingStateByLanguage.has(langServer.language) &&
            this.state.pendingStateByLanguage.get(langServer.language) === 'updating'
        return (
            <div className="site-admin-lang-servers__actions btn-group" role="group">
                {updating && (
                    <span className="site-admin-lang-servers__actions-updating">Pulling latest Docker image…</span>
                )}
                {langServer.state !== 'LANG_SERVER_STATE_DISABLED' &&
                    langServer.canUpdate && (
                        <button
                            disabled={disabled}
                            type="button"
                            className="site-admin-lang-servers__actions-update btn btn-sm"
                            data-tooltip={!disabled ? 'Update server' : undefined}
                            // tslint:disable-next-line:jsx-no-lambda
                            onClick={() => this.updateButtonClicks.next(langServer)}
                        >
                            <DownloadSimpleIcon className="icon-inline" />
                        </button>
                    )}
                {langServer.state !== 'LANG_SERVER_STATE_DISABLED' &&
                    langServer.canRestart && (
                        <button
                            disabled={disabled}
                            type="button"
                            className="site-admin-lang-servers__actions-restart btn btn-secondary"
                            data-tooltip={!disabled ? 'Restart server' : undefined}
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
            <div className="site-admin-lang-servers__repo">
                {langServer.homepageURL &&
                    langServer.homepageURL.startsWith('https://github.com') && (
                        <>
                            <GitHubIcon className="icon-inline" />{' '}
                            <a
                                className="site-admin-lang-servers__repo-link"
                                href={langServer.homepageURL}
                                target="_blank"
                                onClick={this.logClick('LangServerHomepageClicked', langServer)}
                            >
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
                            onClick={this.logClick('LangServerHomepageClicked', langServer)}
                        >
                            {langServer.homepageURL}
                        </a>
                    )}
                {langServer.docsURL && (
                    <a
                        className="site-admin-lang-servers__repo-link"
                        href={langServer.docsURL}
                        target="_blank"
                        data-tooltip="View documentation"
                        onClick={this.logClick('LangServerDocsClicked', langServer)}
                    >
                        <FileDocumentBoxIcon className="icon-inline" />
                    </a>
                )}
                {langServer.issuesURL && (
                    <a
                        className="site-admin-lang-servers__repo-link"
                        href={langServer.issuesURL}
                        target="_blank"
                        data-tooltip="View issues"
                        onClick={this.logClick('LangServerIssuesClicked', langServer)}
                    >
                        <BugIcon className="icon-inline" />
                    </a>
                )}
            </div>
        )
    }

    private logClick(eventLabel: string, langServer: GQL.ILangServer): () => void {
        return () => {
            eventLogger.log(eventLabel, {
                // 🚨 PRIVACY: never provide any private data in { lang_server: { ... } }.
                lang_server: {
                    language: langServer.language,
                },
            })
        }
    }
}
