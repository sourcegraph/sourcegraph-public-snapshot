import Loader from '@sourcegraph/icons/lib/Loader'
import * as H from 'history'
import { upperFirst } from 'lodash'
import * as _monaco from 'monaco-editor' // type only
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { from as fromPromise, Subject, Subscription } from 'rxjs'
import { catchError, concatMap, delay, mergeMap, retryWhen, tap, timeout } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import { SaveToolbar } from '../components/SaveToolbar'
import * as _monacoSettingsEditorModule from '../settings/MonacoSettingsEditor' // type only
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { fetchSite, reloadSite, updateSiteConfiguration } from './backend'
import { siteConfigActions } from './configHelpers'

/**
 * Converts a Monaco/vscode style Disposable object to a simple function that can be added to a rxjs Subscription
 */
const disposableToFn = (disposable: _monaco.IDisposable) => () => disposable.dispose()

interface Props extends RouteComponentProps<any> {
    isLightTheme: boolean
}

interface State {
    site?: GQL.ISite
    loading: boolean
    error?: Error

    /**
     * The contents of the editor in this component.
     */
    contents?: string

    saving?: boolean
    restartToApply: boolean
    reloadStartedAt?: number

    /** The dynamically imported MonacoSettingsEditor module, undefined while loading. */
    monacoSettingsEditorOrError?: typeof _monacoSettingsEditorModule | ErrorLike
}

const EXPECTED_RELOAD_WAIT = 4 * 1000 // 4 seconds

/**
 * A page displaying the site configuration.
 */
export class SiteAdminConfigurationPage extends React.Component<Props, State> {
    public state: State = {
        loading: true,
        restartToApply: window.context.needServerRestart,
    }

    private remoteRefreshes = new Subject<void>()
    private remoteUpdates = new Subject<string>()
    private siteReloads = new Subject<void>()
    private subscriptions = new Subscription()

    private monaco: typeof _monaco | null = null
    private configEditor?: _monaco.editor.ICodeEditor

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminConfiguration')

        this.subscriptions.add(
            fromPromise(import('../settings/MonacoSettingsEditor'))
                .pipe(
                    catchError(error => {
                        console.error(error)
                        return [asError(error)]
                    })
                )
                .subscribe(m => {
                    this.setState({ monacoSettingsEditorOrError: m })
                })
        )

        // Prevent navigation when dirty.
        this.subscriptions.add(
            this.props.history.block((location: H.Location, action: H.Action) => {
                if (action === 'REPLACE') {
                    return undefined
                }
                if (this.state.saving || this.localDirty) {
                    return 'Discard configuration changes?'
                }
                return undefined // allow navigation
            })
        )

        this.subscriptions.add(
            this.remoteRefreshes.pipe(mergeMap(() => fetchSite())).subscribe(
                site =>
                    this.setState({
                        site,
                        error: undefined,
                        loading: false,
                    }),
                error => this.setState({ error, loading: false })
            )
        )
        this.remoteRefreshes.next()

        this.subscriptions.add(
            this.remoteUpdates
                .pipe(
                    tap(() => this.setState({ saving: true, error: undefined })),
                    concatMap(event =>
                        updateSiteConfiguration(event).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ saving: false, error })
                                return []
                            })
                        )
                    ),
                    tap(restartToApply => {
                        if (restartToApply) {
                            window.context.needServerRestart = restartToApply
                        } else {
                            // Refresh site flags so that global site alerts
                            // reflect the latest configuration.
                            refreshSiteFlags().subscribe(undefined, err => console.error(err))
                        }
                        this.setState({ restartToApply })
                        this.remoteRefreshes.next()
                    })
                )
                .subscribe(() => this.setState({ saving: false }), error => this.setState({ saving: false, error }))
        )

        this.subscriptions.add(
            this.siteReloads
                .pipe(
                    tap(() => this.setState({ reloadStartedAt: Date.now(), error: undefined })),
                    mergeMap(reloadSite),
                    delay(2000),
                    mergeMap(() =>
                        // wait for server to restart
                        fetchSite().pipe(
                            retryWhen(x => x.pipe(tap(() => this.forceUpdate()), delay(500))),
                            timeout(5000)
                        )
                    ),
                    tap(() => this.remoteRefreshes.next())
                )
                .subscribe(
                    () => {
                        this.setState({ reloadStartedAt: undefined })
                        window.location.reload() // brute force way to reload view state
                    },
                    error => this.setState({ reloadStartedAt: undefined, error })
                )
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private get remoteContents(): string | undefined {
        return (
            this.state.site &&
            (this.state.site.configuration.pendingContents || this.state.site.configuration.effectiveContents)
        )
    }

    private get localContents(): string | undefined {
        return this.state.contents === undefined ? this.remoteContents : this.state.contents
    }

    private get localDirty(): boolean {
        return !!this.state.site && !!this.state.site.configuration && this.localContents !== this.remoteContents
    }

    public render(): JSX.Element | null {
        const isReloading = typeof this.state.reloadStartedAt === 'number'
        const localContents = this.localContents
        const localDirty = this.localDirty

        const alerts: JSX.Element[] = []
        if (this.state.error) {
            alerts.push(
                <div key="error" className="alert alert-danger site-admin-configuration-page__alert">
                    <p>Error: {this.state.error.message}</p>
                </div>
            )
        }
        if (this.state.reloadStartedAt) {
            alerts.push(
                <div key="error" className="alert alert-primary site-admin-configuration-page__alert">
                    <p>
                        <Loader className="icon-inline" /> Waiting for site to reload...
                    </p>
                    {Date.now() - this.state.reloadStartedAt > EXPECTED_RELOAD_WAIT && (
                        <p>
                            <small>It's taking longer than expected. Check the server logs for error messages.</small>
                        </p>
                    )}
                </div>
            )
        }
        if (this.state.restartToApply) {
            alerts.push(
                <div
                    key="remote-dirty"
                    className="alert alert-warning site-admin-configuration-page__alert site-admin-configuration-page__alert-flex"
                >
                    Server restart is required for the configuration to take effect.
                    <button className="btn btn-primary btn-sm" onClick={this.reloadSite}>
                        Restart server
                    </button>
                </div>
            )
        }
        if (this.state.site && this.state.site.configuration && !this.state.site.configuration.canUpdate) {
            alerts.push(
                <div key="volatile" className="alert alert-danger site-admin-configuration-page__alert">
                    <p>
                        Use this editor as a scratch area for composing Sourcegraph site configuration.{' '}
                        <strong>Changes will NOT be saved</strong> and will be lost when you leave this page.
                    </p>
                    <p>
                        <small>
                            To save and apply changes, manually update{' '}
                            {formatEnvVar(this.state.site.configuration.source)} (in Sourcegraph Server) or{' '}
                            <code>config.json</code> (in Sourcegraph Data Center) with the configuration below and
                            restart the server. Online configuration editing is only supported for Sourcegraph Server
                            and when the configuration lives in a writable file on disk.
                        </small>
                    </p>
                </div>
            )
        }
        if (
            this.state.site &&
            this.state.site.configuration &&
            this.state.site.configuration.validationMessages &&
            this.state.site.configuration.validationMessages.length > 0
        ) {
            alerts.push(
                <div key="validation-messages" className="alert alert-danger site-admin-configuration-page__alert">
                    <p>The server reported issues in the last-saved config:</p>
                    <ul>
                        {this.state.site.configuration.validationMessages.map((e, i) => (
                            <li key={i} className="site-admin-configuration-page__alert-item">
                                {e}
                            </li>
                        ))}
                    </ul>
                </div>
            )
        }

        // Avoid user confusion on Data Center config.
        //
        // To get a list of all keys: jq '.properties | keys' < datacenter.schema.json
        const dataCenterProps = [
            'alertmanagerConfig',
            'alertmanagerURL',
            'authProxyIP',
            'authProxyPassword',
            'deploymentOverrides',
            'gitoliteIP',
            'gitserverCount',
            'gitserverDiskSize',
            'gitserverSSH',
            'httpNodePort',
            'httpsNodePort',
            'indexedSearchDiskSize',
            'langGo',
            'langJava',
            'langJavaScript',
            'langPHP',
            'langPython',
            'langSwift',
            'langTypeScript',
            'nodeSSDPath',
            'phabricatorIP',
            'prometheus',
            'pyPIIP',
            'rbac',
            'storageClass',
            'useAlertManager',
        ].filter(prop => localContents && localContents.includes(`"${prop}"`))
        if (dataCenterProps.length > 0) {
            alerts.push(
                <div key="datacenter-props-present" className="alert alert-info site-admin-configuration-page__alert">
                    The configuration contains properties that are valid only in Sourcegraph Data Center's{' '}
                    <code>config.json</code> file: <code>{dataCenterProps.join(' ')}</code>. You can disregard the
                    validation warnings for these properties reported by the configuration editor.
                </div>
            )
        }

        return (
            <div className="site-admin-configuration-page">
                <PageTitle title="Configuration - Admin" />
                <h2>Site configuration</h2>
                <p>
                    View and edit the Sourcegraph site configuration. See{' '}
                    <a href="https://about.sourcegraph.com/docs/server/">documentation</a> for more information.
                </p>
                <div className="site-admin-configuration-page__alerts">{alerts}</div>
                {this.state.loading && <Loader className="icon-inline" />}
                {this.state.site &&
                    this.state.site.configuration && (
                        <div>
                            {this.state.monacoSettingsEditorOrError === undefined ? (
                                <Loader className="icon-inline" />
                            ) : isErrorLike(this.state.monacoSettingsEditorOrError) ? (
                                <div className="alert alert-danger">
                                    Error loading site configuration editor:{' '}
                                    {upperFirst(this.state.monacoSettingsEditorOrError.message)}
                                </div>
                            ) : (
                                (() => {
                                    const MonacoSettingsEditor = this.state.monacoSettingsEditorOrError
                                        .MonacoSettingsEditor
                                    return (
                                        <>
                                            <div className="site-admin-configuration-page__action-groups">
                                                <div className="site-admin-configuration-page__action-groups">
                                                    <div className="site-admin-configuration-page__action-group-header">
                                                        Quick configure:
                                                    </div>
                                                    <div className="site-admin-configuration-page__actions">
                                                        {siteConfigActions.map(({ id, label }) => (
                                                            <button
                                                                key={id}
                                                                className="btn btn-secondary btn-sm site-admin-configuration-page__action"
                                                                // tslint:disable-next-line:jsx-no-lambda
                                                                onClick={() => this.runAction(id, this.configEditor)}
                                                            >
                                                                {label}
                                                            </button>
                                                        ))}
                                                    </div>
                                                </div>
                                            </div>
                                            {this.state.site.configuration.canUpdate && (
                                                <SaveToolbar
                                                    dirty={localDirty}
                                                    disabled={isReloading || this.state.saving || !localDirty}
                                                    saving={this.state.saving}
                                                    onSave={this.save}
                                                    onDiscard={this.discard}
                                                />
                                            )}
                                            <MonacoSettingsEditor
                                                className="site-admin-configuration-page__config"
                                                value={localContents}
                                                jsonSchema="https://sourcegraph.com/v1/site.schema.json#"
                                                onChange={this.onDidChange}
                                                readOnly={isReloading || this.state.saving}
                                                height={600}
                                                monacoRef={this.monacoRef}
                                                isLightTheme={this.props.isLightTheme}
                                                onDidSave={this.save}
                                            />
                                        </>
                                    )
                                })()
                            )}
                            <p className="form-text">
                                <small>Source: {formatEnvVar(this.state.site.configuration.source)}</small>
                            </p>
                            <p className="form-text">
                                <small>
                                    Use Ctrl+Space for completion, and hover over JSON properties for documentation. For
                                    more information, see the{' '}
                                    <a href="https://about.sourcegraph.com/docs/server/config/settings">
                                        documentation
                                    </a>.
                                </small>
                            </p>
                        </div>
                    )}
            </div>
        )
    }

    private onDidChange = (newValue: string) => this.setState({ contents: newValue })

    private discard = () => {
        if (
            this.state.contents === undefined ||
            this.remoteContents === this.state.contents ||
            window.confirm('Really discard edits?')
        ) {
            eventLogger.log('SiteConfigurationDiscarded')
            this.setState({ contents: undefined, error: undefined })
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private save = () => {
        eventLogger.log('SiteConfigurationSaved')
        this.remoteUpdates.next(this.state.contents)
    }

    private reloadSite = () => {
        eventLogger.log('SiteReloaded')
        this.siteReloads.next()
    }

    private monacoRef = (monacoValue: typeof _monaco | null) => {
        this.monaco = monacoValue
        // This function can only be called if the editor was loaded so it is okay to cast here
        const monacoSettingsEditor = this.state.monacoSettingsEditorOrError as typeof _monacoSettingsEditorModule
        if (this.monaco && monacoSettingsEditor) {
            this.subscriptions.add(
                disposableToFn(
                    this.monaco.editor.onDidCreateEditor(editor => {
                        this.configEditor = editor
                    })
                )
            )
            this.subscriptions.add(
                disposableToFn(
                    this.monaco.editor.onDidCreateModel(model => {
                        if (this.configEditor && monacoSettingsEditor.isStandaloneCodeEditor(this.configEditor)) {
                            for (const { id, label, run } of siteConfigActions) {
                                monacoSettingsEditor.addEditorAction(this.configEditor, model, label, id, run)
                            }
                        }
                    })
                )
            )
        }
    }

    private runAction(id: string, editor?: _monaco.editor.ICodeEditor): void {
        if (editor) {
            const action = editor.getAction(id)
            action.run().done(() => void 0, (err: any) => console.error(err))
        } else {
            alert('Wait for editor to load before running action.')
        }
    }
}

function formatEnvVar(text: string): React.ReactChild[] | string {
    const S = 'SOURCEGRAPH_CONFIG'
    const idx = text.indexOf(S)
    if (idx === -1) {
        return text
    }
    return [text.slice(0, idx), <code key={S}>{S}</code>, text.slice(idx + S.length)]
}
