import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import { upperFirst } from 'lodash'
import * as _monaco from 'monaco-editor' // type only
import * as React from 'react'
import { from as fromPromise, Subscription } from 'rxjs'
import { catchError } from 'rxjs/operators'
import { asError, ErrorLike, isErrorLike } from '../../../shared/src/util/errors'
import { SaveToolbar } from '../components/SaveToolbar'
import * as _monacoSettingsEditorModule from '../settings/MonacoSettingsEditor' // type only
import { EditorAction } from '../site-admin/configHelpers'

/**
 * Converts a Monaco/vscode style Disposable object to a simple function that can be added to a rxjs Subscription
 */
const disposableToFn = (disposable: _monaco.IDisposable) => () => disposable.dispose()

interface Props
    extends Pick<
            _monacoSettingsEditorModule.Props,
            'id' | 'readOnly' | 'height' | 'jsonSchemaId' | 'extraSchemas' | 'isLightTheme'
        > {
    value: string

    actions?: EditorAction[]

    loading?: boolean
    saving?: boolean

    canEdit?: boolean

    onSave?: (value: string) => void
    onChange?: (value: string) => void
    onDirtyChange?: (dirty: boolean) => void

    isLightTheme: boolean

    history: H.History
}

interface State {
    /** The current contents of the editor, if changed from Props.value. */
    value?: string

    /** The dynamically imported MonacoSettingsEditor module, undefined while loading. */
    monacoSettingsEditorOrError?: typeof _monacoSettingsEditorModule | ErrorLike
}

/** Displays a MonacoSettingsEditor component without loading Monaco in the current Webpack chunk. */
export class DynamicallyImportedMonacoSettingsEditor extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    private monaco: typeof _monaco | null = null
    private configEditor?: _monaco.editor.ICodeEditor

    public componentDidMount(): void {
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
                if (this.props.loading || this.isDirty) {
                    return 'Discard changes?'
                }
                return undefined // allow navigation
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private get effectiveValue(): string {
        return this.state.value === undefined ? this.props.value : this.state.value
    }

    private get isDirty(): boolean {
        return this.effectiveValue !== this.props.value
    }

    public render(): JSX.Element | null {
        if (this.state.monacoSettingsEditorOrError === undefined) {
            return <LoadingSpinner className="icon-inline" />
        }

        if (isErrorLike(this.state.monacoSettingsEditorOrError)) {
            return (
                <div className="alert alert-danger">
                    Error loading: {upperFirst(this.state.monacoSettingsEditorOrError.message)}
                </div>
            )
        }

        const isDirty = this.isDirty
        const effectiveValue = this.effectiveValue

        const MonacoSettingsEditor = this.state.monacoSettingsEditorOrError.MonacoSettingsEditor
        return (
            <>
                {this.props.actions && (
                    <div className="site-admin-configuration-page__action-groups">
                        <div className="site-admin-configuration-page__action-groups">
                            <div className="site-admin-configuration-page__action-group-header">Quick configure:</div>
                            <div className="site-admin-configuration-page__actions">
                                {this.props.actions.map(({ id, label }) => (
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
                )}
                {this.props.canEdit && (
                    <SaveToolbar
                        dirty={isDirty}
                        disabled={this.props.loading || this.props.saving || !isDirty}
                        saving={this.props.saving}
                        onSave={this.onSave}
                        onDiscard={this.discard}
                    />
                )}
                <MonacoSettingsEditor
                    {...this.props}
                    onDidSave={this.onSave}
                    onChange={this.onChange}
                    value={effectiveValue}
                    monacoRef={this.monacoRef}
                />
            </>
        )
    }

    private onSave = () => {
        const value = this.effectiveValue
        if (this.props.onSave) {
            this.props.onSave(value)
        }
    }

    private onChange = (newValue: string) => {
        this.setState({ value: newValue }, () => {
            if (this.props.onChange) {
                this.props.onChange(newValue)
            }
            if (this.props.onDirtyChange) {
                this.props.onDirtyChange(this.isDirty)
            }
        })
    }

    private discard = () => {
        if (
            this.state.value === undefined ||
            this.props.value === this.state.value ||
            window.confirm('Discard edits?')
        ) {
            this.setState({ value: undefined })
        }
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
                        if (
                            this.configEditor &&
                            monacoSettingsEditor.isStandaloneCodeEditor(this.configEditor) &&
                            this.props.actions
                        ) {
                            for (const { id, label, run } of this.props.actions) {
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
            action.run().then(() => void 0, (err: any) => console.error(err))
        } else {
            alert('Wait for editor to load before running action.')
        }
    }
}
