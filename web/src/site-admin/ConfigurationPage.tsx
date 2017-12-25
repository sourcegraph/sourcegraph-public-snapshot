import Loader from '@sourcegraph/icons/lib/Loader'
import { applyEdits } from '@sqs/jsonc-parser/lib/format'
import * as H from 'history'
import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { delay } from 'rxjs/operators/delay'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { retryWhen } from 'rxjs/operators/retryWhen'
import { tap } from 'rxjs/operators/tap'
import { timeout } from 'rxjs/operators/timeout'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { PageTitle } from '../components/PageTitle'
import { SaveToolbar } from '../components/SaveToolbar'
import { isStandaloneCodeEditor, MonacoSettingsEditor, toMonacoEdits } from '../settings/MonacoSettingsEditor'
import { eventLogger } from '../tracking/eventLogger'
import { fetchSite, reloadSite, updateSiteConfiguration } from './backend'
import { editorActions } from './configHelpers'

interface Props extends RouteComponentProps<any> {}

interface State {
    site?: GQL.ISite
    error?: Error

    /**
     * The contents of the editor in this component.
     */
    contents?: string

    saving?: boolean
    reloadStartedAt?: number
}

const EXPECTED_RELOAD_WAIT = 4 * 1000 // 4 seconds

/**
 * A page displaying the site configuration.
 */
export class ConfigurationPage extends React.Component<Props, State> {
    public state: State = {}

    private remoteRefreshes = new Subject<void>()
    private remoteUpdates = new Subject<string>()
    private siteReloads = new Subject<void>()
    private subscriptions = new Subscription()

    private monaco: typeof monaco | null
    private editor: monaco.editor.ICodeEditor

    public componentDidMount(): void {
        eventLogger.logViewEvent('SiteAdminConfiguration')

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
            this.remoteRefreshes.pipe(mergeMap(fetchSite)).subscribe(
                site =>
                    this.setState({
                        site,
                        error: undefined,
                    }),
                error => this.setState({ error })
            )
        )
        this.remoteRefreshes.next()

        this.subscriptions.add(
            this.remoteUpdates
                .pipe(
                    tap(() => this.setState({ saving: true, error: undefined })),
                    mergeMap(updateSiteConfiguration),
                    tap(() => this.remoteRefreshes.next())
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
        const remoteDirty =
            this.state.site &&
            !this.state.reloadStartedAt &&
            this.state.site.configuration &&
            typeof this.state.site.configuration.pendingContents === 'string'
        const localDirty = this.localDirty

        const alerts: JSX.Element[] = []
        if (this.state.error) {
            alerts.push(
                <div key="error" className="alert alert-danger site-admin-configuration-page__alert">
                    <p>Error: {this.state.error}</p>
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
        if (remoteDirty) {
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
        if (
            this.state.site &&
            this.state.site.configuration &&
            !this.state.site.configuration.canUpdate &&
            localDirty
        ) {
            alerts.push(
                <div key="volatile" className="alert alert-danger site-admin-configuration-page__alert">
                    <p>
                        <strong>Changes will NOT be saved</strong> and will be lost when you leave this page.
                    </p>
                    <p>
                        <small>
                            To save and apply changes, manually update{' '}
                            {formatEnvVar(this.state.site.configuration.source)} with the configuration below and
                            restart the server. Online configuration editing is only supported when the configuration
                            lives in a writable file on disk.
                        </small>
                    </p>
                </div>
            )
        }

        return (
            <div className="site-admin-configuration-page">
                <PageTitle title="Configuration - Admin" />
                <h2>Configuration</h2>
                <p>
                    View and edit the Sourcegraph site configuration below. See{' '}
                    <a href="https://about.sourcegraph.com/docs/server/">documentation</a> for more information.
                </p>
                <div className="site-admin-configuration-page__alerts">{alerts}</div>
                {this.state.site &&
                    this.state.site.configuration && (
                        <div>
                            <div className="site-admin-configuration-page__editor-actions">
                                {editorActions.map(({ id, label }) => (
                                    <button
                                        key={id}
                                        className="btn btn-primary btn-sm site-admin-configuration-page__editor-action"
                                        // tslint:disable-next-line:jsx-no-lambda
                                        onClick={() => this.runAction(id)}
                                    >
                                        {label}
                                    </button>
                                ))}
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
                            />
                            <p className="form-text">
                                <small>Source: {formatEnvVar(this.state.site.configuration.source)}</small>
                            </p>
                            <p className="form-text">
                                <small>
                                    Use Cmd/Ctrl+Space for completion, and hover over JSON properties for documentation.
                                    For more information, see the{' '}
                                    <a href="https://about.sourcegraph.com/docs/server/">documentation</a>.
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
            this.setState({ contents: undefined })
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private save = () => {
        eventLogger.log('SiteConfigurationSaved')
        this.remoteUpdates.next(this.state.contents)
    }

    private reloadSite = () => this.siteReloads.next()

    private monacoRef = (monacoValue: typeof monaco | null) => {
        this.monaco = monacoValue
        if (this.monaco) {
            this.subscriptions.add(this.monaco.editor.onDidCreateEditor(editor => (this.editor = editor)).dispose)
            this.subscriptions.add(
                this.monaco.editor.onDidCreateModel(model => {
                    if (this.editor && isStandaloneCodeEditor(this.editor)) {
                        for (const { id, label, run } of editorActions) {
                            this.editor.addAction({
                                label,
                                id,
                                run: editor => {
                                    editor.focus()
                                    editor.pushUndoStop()
                                    const { edits, selectText } = run(editor.getValue())
                                    const monacoEdits = toMonacoEdits(model, edits)
                                    let selection: monaco.Selection | undefined
                                    if (typeof selectText === 'string') {
                                        const afterText = applyEdits(editor.getValue(), edits)
                                        let offset = afterText.slice(edits[0].offset).indexOf(selectText)
                                        if (offset !== -1) {
                                            offset += edits[0].offset
                                            selection = monaco.Selection.fromPositions(
                                                getPositionAt(afterText, offset),
                                                getPositionAt(afterText, offset + selectText.length)
                                            )
                                        }
                                    }
                                    if (!selection) {
                                        selection = monaco.Selection.fromPositions(
                                            monacoEdits[0].range.getStartPosition(),
                                            monacoEdits[monacoEdits.length - 1].range.getEndPosition()
                                        )
                                    }
                                    editor.executeEdits(id, monacoEdits, [selection])
                                    editor.revealPositionInCenter(selection.getStartPosition())
                                },
                            })
                        }
                    }
                }).dispose
            )
        }
    }

    private runAction(id: string): void {
        if (this.editor) {
            const action = this.editor.getAction(id)
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

function getPositionAt(text: string, offset: number): monaco.Position {
    const lines = text.split('\n')
    let pos = 0
    for (const [i, line] of lines.entries()) {
        if (offset < pos + line.length + 1) {
            return new monaco.Position(i + 1, offset - pos + 1)
        }
        pos += line.length + 1
    }
    throw new Error(`offset ${offset} out of bounds in text of length ${text.length}`)
}
