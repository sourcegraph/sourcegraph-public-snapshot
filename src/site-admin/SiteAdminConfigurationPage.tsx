import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { Subject, Subscription } from 'rxjs'
import { catchError, concatMap, delay, mergeMap, retryWhen, tap, timeout } from 'rxjs/operators'
import * as GQL from '../backend/graphqlschema'
import { PageTitle } from '../components/PageTitle'
import siteSchemaJSON from '../schema/site.schema.json'
import { DynamicallyImportedMonacoSettingsEditor } from '../settings/DynamicallyImportedMonacoSettingsEditor'
import { refreshSiteFlags } from '../site/backend'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite, reloadSite, updateSiteConfiguration } from './backend'
import { siteConfigActions } from './configHelpers'

interface Props extends RouteComponentProps<any> {
    isLightTheme: boolean
}

interface State {
    site?: GQL.ISite
    loading: boolean
    error?: Error

    isDirty?: boolean
    saving?: boolean
    restartToApply: boolean
    reloadStartedAt?: number
}

const EXPECTED_RELOAD_WAIT = 7 * 1000 // 7 seconds

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

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminConfiguration')

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
                            retryWhen(x =>
                                x.pipe(
                                    tap(() => this.forceUpdate()),
                                    delay(500)
                                )
                            ),
                            timeout(10000)
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

    public render(): JSX.Element | null {
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
                        <LoadingSpinner className="icon-inline" /> Waiting for site to reload...
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
                            {formatEnvVar(this.state.site.configuration.source)} (or <code>values.yaml</code> for
                            Kubernetes cluster deployments) with the configuration below and restart the server. Online
                            configuration editing is only supported when deployed on Docker and when the configuration
                            lives in a writable file on disk.
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

        // Avoid user confusion with values.yaml properties mixed in with site config properties.
        const contents =
            this.state.site && this.state.site.configuration && this.state.site.configuration.effectiveContents
        const legacyKubernetesConfigProps = [
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
        ].filter(prop => contents && contents.includes(`"${prop}"`))
        if (legacyKubernetesConfigProps.length > 0) {
            alerts.push(
                <div
                    key="legacy-cluster-props-present"
                    className="alert alert-info site-admin-configuration-page__alert"
                >
                    The configuration contains properties that are valid only in the
                    <code>values.yaml</code> config file used for Kubernetes cluster deployments of Sourcegraph:{' '}
                    <code>{legacyKubernetesConfigProps.join(' ')}</code>. You can disregard the validation warnings for
                    these properties reported by the configuration editor.
                </div>
            )
        }

        const isReloading = typeof this.state.reloadStartedAt === 'number'

        return (
            <div className="site-admin-configuration-page">
                <PageTitle title="Configuration - Admin" />
                <h2>Site configuration</h2>
                <p>
                    View and edit the Sourcegraph site configuration. See{' '}
                    <a href="https://docs.sourcegraph.com/admin/site_config">documentation</a> for more information.
                </p>
                <div className="site-admin-configuration-page__alerts">{alerts}</div>
                {this.state.loading && <LoadingSpinner className="icon-inline" />}
                {this.state.site &&
                    this.state.site.configuration && (
                        <div>
                            <DynamicallyImportedMonacoSettingsEditor
                                value={contents || ''}
                                actions={siteConfigActions}
                                jsonSchema={siteSchemaJSON}
                                onDirtyChange={this.onDirtyChange}
                                canEdit={this.state.site.configuration.canUpdate}
                                saving={this.state.saving}
                                loading={isReloading || this.state.saving}
                                height={600}
                                isLightTheme={this.props.isLightTheme}
                                onSave={this.onSave}
                                history={this.props.history}
                            />
                            <p className="form-text text-muted">
                                <small>Source: {formatEnvVar(this.state.site.configuration.source)}</small>
                            </p>
                            <p className="form-text text-muted">
                                <small>
                                    Use Ctrl+Space for completion, and hover over JSON properties for documentation. For
                                    more information, see the{' '}
                                    <a href="https://docs.sourcegraph.com/admin/site_config/all">documentation</a>.
                                </small>
                            </p>
                        </div>
                    )}
            </div>
        )
    }

    private onDirtyChange = (isDirty: boolean) => this.setState({ isDirty })

    private onSave = (value: string) => {
        eventLogger.log('SiteConfigurationSaved')
        this.remoteUpdates.next(value)
    }

    private reloadSite = () => {
        eventLogger.log('SiteReloaded')
        this.siteReloads.next()
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
