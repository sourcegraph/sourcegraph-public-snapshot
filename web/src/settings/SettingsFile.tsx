import * as H from 'history'
import * as React from 'react'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { SaveToolbar } from '../components/SaveToolbar'
import { settingsActions } from '../site-admin/configHelpers'
import { eventLogger } from '../tracking/eventLogger'
import { addEditorAction } from '../util/monaco'
import { isStandaloneCodeEditor, MonacoSettingsEditor } from './MonacoSettingsEditor'

interface Props {
    history: H.History

    settings: GQL.ISettings | null

    /**
     * Called when the user saves changes to the settings file's contents.
     */
    onDidCommit: (lastKnownSettingsID: number | null, contents: string) => void

    /**
     * Called when the user discards changes to the settings file's contents.
     */
    onDidDiscard: () => void

    /**
     * The error that occurred on the last call to the onDidCommit callback,
     * if any.
     */
    commitError?: Error

    isLightTheme: boolean
}

interface State {
    contents?: string
    saving: boolean

    /**
     * The lastKnownSettingsID that we started editing from. If null, then no
     * previous versions of the settings exist, and we're creating them from
     * scratch.
     */
    editingLastKnownSettingsID?: number | null
}

const emptySettings = '{\n  // add settings here (Cmd/Ctrl+Space to see hints)\n}'

const disposableToFn = (disposable: monaco.IDisposable) => () => disposable.dispose()

export class SettingsFile extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private editor?: monaco.editor.ICodeEditor
    private monaco: typeof monaco | null = null

    constructor(props: Props) {
        super(props)

        this.state = { saving: false }

        // Reset state upon navigation to a different subject.
        this.componentUpdates
            .pipe(startWith(props), map(({ settings }) => settings), distinctUntilChanged())
            .subscribe(settings => {
                if (this.state.contents !== undefined) {
                    this.setState({ contents: undefined })
                }
            })

        // Saving ended (in failure) if we get a commitError.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ commitError }) => commitError),
                    distinctUntilChanged(),
                    filter(commitError => !!commitError)
                )
                .subscribe(() => this.setState({ saving: false }))
        )

        // We are finished saving when we receive the new settings ID and it's
        // higher than the one we saved on top of.
        const refreshedAfterSave = this.componentUpdates.pipe(
            filter(({ settings }) => !!settings),
            distinctUntilChanged(
                (a, b) =>
                    (!a.settings && !!b.settings) ||
                    (!!a.settings && !b.settings) ||
                    (!!a.settings &&
                        !!b.settings &&
                        a.settings.configuration.contents === b.settings.configuration.contents &&
                        a.settings.id === b.settings.id)
            ),
            filter(
                ({ settings, commitError }) =>
                    !!settings &&
                    !commitError &&
                    ((typeof this.state.editingLastKnownSettingsID === 'number' &&
                        settings.id > this.state.editingLastKnownSettingsID) ||
                        (typeof settings.id === 'number' && this.state.editingLastKnownSettingsID === null))
            )
        )
        this.subscriptions.add(
            refreshedAfterSave.subscribe(({ settings }) =>
                this.setState({
                    saving: false,
                    editingLastKnownSettingsID: undefined,
                    contents: settings ? settings.configuration.contents : undefined,
                })
            )
        )
    }

    public componentDidMount(): void {
        // Prevent navigation when dirty.
        this.subscriptions.add(
            this.props.history.block((location: H.Location, action: H.Action) => {
                if (action === 'REPLACE') {
                    return undefined
                }
                if (this.state.saving || this.dirty) {
                    return 'Discard settings changes?'
                }
                return undefined // allow navigation
            })
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private get dirty(): boolean {
        return this.state.contents !== undefined && this.state.contents !== this.getPropsSettingsContentsOrEmpty()
    }

    public render(): JSX.Element | null {
        const dirty = this.dirty
        const contents =
            this.state.contents === undefined ? this.getPropsSettingsContentsOrEmpty() : this.state.contents

        return (
            <div className="settings-file">
                <div className="site-admin-configuration-page__action-groups">
                    <div className="site-admin-configuration-page__action-groups">
                        <div className="site-admin-configuration-page__action-group-header">Quick configure:</div>
                        <div className="site-admin-configuration-page__actions">
                            {settingsActions.map(({ id, label }) => (
                                <button
                                    key={id}
                                    className="btn btn-secondary btn-sm site-admin-configuration-page__action"
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={() => this.runAction(id)}
                                >
                                    {label}
                                </button>
                            ))}
                        </div>
                    </div>
                </div>
                <SaveToolbar
                    dirty={dirty}
                    disabled={this.state.saving || !dirty}
                    error={this.props.commitError}
                    saving={this.state.saving}
                    onSave={this.save}
                    onDiscard={this.discard}
                />
                <MonacoSettingsEditor
                    className="settings-file__contents form-control"
                    value={contents}
                    jsonSchema="https://sourcegraph.com/v1/settings.schema.json#"
                    onChange={this.onEditorChange}
                    readOnly={this.state.saving}
                    monacoRef={this.monacoRef}
                    isLightTheme={this.props.isLightTheme}
                    onDidSave={this.save}
                />
            </div>
        )
    }

    private monacoRef = (monacoValue: typeof monaco | null) => {
        this.monaco = monacoValue
        if (this.monaco) {
            this.subscriptions.add(
                disposableToFn(
                    this.monaco.editor.onDidCreateEditor(editor => {
                        this.editor = editor
                    })
                )
            )
            this.subscriptions.add(
                disposableToFn(
                    this.monaco.editor.onDidCreateModel(model => {
                        if (this.editor && isStandaloneCodeEditor(this.editor)) {
                            for (const { id, label, run } of settingsActions) {
                                addEditorAction(this.editor, model, label, id, run)
                            }
                        }
                    })
                )
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

    private getPropsSettingsContentsOrEmpty(settings = this.props.settings): string {
        return settings ? settings.configuration.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return this.props.settings ? this.props.settings.id : null
    }

    private discard = () => {
        if (this.getPropsSettingsContentsOrEmpty() === this.state.contents || window.confirm('Really discard edits?')) {
            eventLogger.log('SettingsFileDiscard')
            this.setState({
                contents: undefined,
                editingLastKnownSettingsID: undefined,
            })
            this.props.onDidDiscard()
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private onEditorChange = (newValue: string) => {
        if (newValue !== this.getPropsSettingsContentsOrEmpty()) {
            this.setState({ editingLastKnownSettingsID: this.getPropsSettingsID() })
        }
        this.setState({ contents: newValue })
    }

    private save = () => {
        eventLogger.log('SettingsFileSaved')
        this.setState({ saving: true }, () => {
            this.props.onDidCommit(this.getPropsSettingsID(), this.state.contents!)
        })
    }
}
