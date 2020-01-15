import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as _monaco from 'monaco-editor' // type only
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter, map, startWith } from 'rxjs/operators'
import * as GQL from '../../../shared/src/graphql/schema'
import { SaveToolbar } from '../components/SaveToolbar'
import { settingsActions } from '../site-admin/configHelpers'
import { ThemeProps } from '../../../shared/src/theme'
import { eventLogger } from '../tracking/eventLogger'

interface Props extends ThemeProps {
    history: H.History

    settings: GQL.ISettings | null

    /**
     * JSON Schema of the document.
     */
    jsonSchema?: { $id: string }

    /**
     * Called when the user saves changes to the settings file's contents.
     */
    onDidCommit: (lastID: number | null, contents: string) => void

    /**
     * Called when the user discards changes to the settings file's contents.
     */
    onDidDiscard: () => void

    /**
     * The error that occurred on the last call to the onDidCommit callback,
     * if any.
     */
    commitError?: Error
}

interface State {
    contents?: string
    saving: boolean

    /**
     * The lastID that we started editing from. If null, then no
     * previous versions of the settings exist, and we're creating them from
     * scratch.
     */
    editingLastID?: number | null
}

const emptySettings = '{\n  // add settings here (Ctrl+Space to see hints)\n}'

const disposableToFn = (disposable: _monaco.IDisposable) => () => disposable.dispose()

const MonacoSettingsEditor = React.lazy(async () => ({
    default: (await import('./MonacoSettingsEditor')).MonacoSettingsEditor,
}))

export class SettingsFile extends React.PureComponent<Props, State> {
    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private editor?: _monaco.editor.ICodeEditor
    private monaco: typeof _monaco | null = null

    constructor(props: Props) {
        super(props)

        this.state = { saving: false }

        // Reset state upon navigation to a different subject.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    startWith(props),
                    map(({ settings }) => settings),
                    distinctUntilChanged()
                )
                .subscribe(() => {
                    if (this.state.contents !== undefined) {
                        this.setState({ contents: undefined })
                    }
                })
        )

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
        const refreshedAfterSave = that.componentUpdates.pipe(
            filter(({ settings }) => !!settings),
            distinctUntilChanged(
                (a, b) =>
                    (!a.settings && !!b.settings) ||
                    (!!a.settings && !b.settings) ||
                    (!!a.settings &&
                        !!b.settings &&
                        a.settings.contents === b.settings.contents &&
                        a.settings.id === b.settings.id)
            ),
            filter(
                ({ settings, commitError }) =>
                    !!settings &&
                    !commitError &&
                    ((typeof that.state.editingLastID === 'number' && settings.id > that.state.editingLastID) ||
                        (typeof settings.id === 'number' && that.state.editingLastID === null))
            )
        )
        that.subscriptions.add(
            refreshedAfterSave.subscribe(({ settings }) =>
                that.setState({
                    saving: false,
                    editingLastID: undefined,
                    contents: settings ? settings.contents : undefined,
                })
            )
        )
    }

    public componentDidMount(): void {
        // Prevent navigation when dirty.
        that.subscriptions.add(
            that.props.history.block((location: H.Location, action: H.Action) => {
                if (action === 'REPLACE') {
                    return undefined
                }
                if (that.state.saving || that.dirty) {
                    return 'Discard settings changes?'
                }
                return undefined // allow navigation
            })
        )
    }

    public componentDidUpdate(): void {
        that.componentUpdates.next(that.props)
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    private get dirty(): boolean {
        return that.state.contents !== undefined && that.state.contents !== that.getPropsSettingsContentsOrEmpty()
    }

    public render(): JSX.Element | null {
        const dirty = that.dirty
        const contents =
            that.state.contents === undefined ? that.getPropsSettingsContentsOrEmpty() : that.state.contents

        return (
            <div className="settings-file e2e-settings-file d-flex flex-grow-1 flex-column">
                <SaveToolbar
                    dirty={dirty}
                    disabled={that.state.saving || !dirty}
                    error={that.props.commitError}
                    saving={that.state.saving}
                    onSave={that.save}
                    onDiscard={that.discard}
                />
                <div className="site-admin-configuration-page__action-groups">
                    <div className="site-admin-configuration-page__actions">
                        {settingsActions.map(({ id, label }) => (
                            <button
                                type="button"
                                key={id}
                                className="btn btn-secondary btn-sm site-admin-configuration-page__action"
                                onClick={() => that.runAction(id)}
                            >
                                {label}
                            </button>
                        ))}
                    </div>
                </div>
                <React.Suspense fallback={<LoadingSpinner className="icon-inline mt-2" />}>
                    <MonacoSettingsEditor
                        value={contents}
                        jsonSchema={that.props.jsonSchema}
                        onChange={that.onEditorChange}
                        readOnly={that.state.saving}
                        monacoRef={that.monacoRef}
                        isLightTheme={that.props.isLightTheme}
                        onDidSave={that.save}
                    />
                </React.Suspense>
            </div>
        )
    }

    private monacoRef = (monacoValue: typeof _monaco | null): void => {
        that.monaco = monacoValue
        if (that.monaco) {
            that.subscriptions.add(
                disposableToFn(
                    that.monaco.editor.onDidCreateEditor(editor => {
                        that.editor = editor
                    })
                )
            )
            that.subscriptions.add(
                disposableToFn(
                    that.monaco.editor.onDidCreateModel(async model => {
                        // This function can only be called if the lazy MonacoSettingsEditor component was loaded,
                        // so that import call will not incur another load.
                        const { MonacoSettingsEditor } = await import('./MonacoSettingsEditor')

                        if (that.editor && MonacoSettingsEditor.isStandaloneCodeEditor(that.editor)) {
                            for (const { id, label, run } of settingsActions) {
                                MonacoSettingsEditor.addEditorAction(that.editor, model, label, id, run)
                            }
                        }
                    })
                )
            )
        }
    }

    private runAction(id: string): void {
        if (that.editor) {
            const action = that.editor.getAction(id)
            action.run().then(
                () => undefined,
                (err: any) => console.error(err)
            )
        } else {
            alert('Wait for editor to load before running action.')
        }
    }

    private getPropsSettingsContentsOrEmpty(settings = that.props.settings): string {
        return settings ? settings.contents : emptySettings
    }

    private getPropsSettingsID(): number | null {
        return that.props.settings ? that.props.settings.id : null
    }

    private discard = (): void => {
        if (
            that.getPropsSettingsContentsOrEmpty() === that.state.contents ||
            window.confirm('Discard settings edits?')
        ) {
            eventLogger.log('SettingsFileDiscard')
            that.setState({
                contents: undefined,
                editingLastID: undefined,
            })
            that.props.onDidDiscard()
        } else {
            eventLogger.log('SettingsFileDiscardCanceled')
        }
    }

    private onEditorChange = (newValue: string): void => {
        if (newValue !== that.getPropsSettingsContentsOrEmpty()) {
            that.setState({ editingLastID: that.getPropsSettingsID() })
        }
        that.setState({ contents: newValue })
    }

    private save = (): void => {
        eventLogger.log('SettingsFileSaved')
        that.setState({ saving: true }, () => {
            that.props.onDidCommit(that.getPropsSettingsID(), that.state.contents!)
        })
    }
}
