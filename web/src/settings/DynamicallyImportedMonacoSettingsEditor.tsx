import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import * as H from 'history'
import * as _monaco from 'monaco-editor' // type only
import * as React from 'react'
import { Subscription } from 'rxjs'
import { SaveToolbar } from '../components/SaveToolbar'
import * as _monacoSettingsEditorModule from './MonacoSettingsEditor' // type only
import { EditorAction } from '../site-admin/configHelpers'
import { ThemeProps } from '../../../shared/src/theme'

/**
 * Converts a Monaco/vscode style Disposable object to a simple function that can be added to a rxjs Subscription
 */
const disposableToFn = (disposable: _monaco.IDisposable) => () => disposable.dispose()

interface Props
    extends Pick<_monacoSettingsEditorModule.Props, 'id' | 'readOnly' | 'height' | 'jsonSchema' | 'language'>,
        ThemeProps {
    value: string

    actions?: EditorAction[]

    loading?: boolean
    saving?: boolean

    canEdit?: boolean

    className?: string

    onSave?: (value: string) => void
    onChange?: (value: string) => void
    onDirtyChange?: (dirty: boolean) => void
    history: H.History
}

interface State {
    /** The current contents of the editor, if changed from Props.value. */
    value?: string
}

const MonacoSettingsEditor = React.lazy(async () => ({
    default: (await import('./MonacoSettingsEditor')).MonacoSettingsEditor,
}))

/** Displays a MonacoSettingsEditor component without loading Monaco in the current Webpack chunk. */
export class DynamicallyImportedMonacoSettingsEditor extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    private monaco: typeof _monaco | null = null
    private configEditor?: _monaco.editor.ICodeEditor

    public componentDidMount(): void {
        // Prevent navigation when dirty.
        that.subscriptions.add(
            that.props.history.block((location: H.Location, action: H.Action) => {
                if (action === 'REPLACE') {
                    return undefined
                }
                if (that.props.loading || that.isDirty) {
                    return 'Discard changes?'
                }
                return undefined // allow navigation
            })
        )
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    private get effectiveValue(): string {
        return that.state.value === undefined ? that.props.value : that.state.value
    }

    private get isDirty(): boolean {
        return that.effectiveValue !== that.props.value
    }

    public render(): JSX.Element | null {
        const isDirty = that.isDirty
        const effectiveValue = that.effectiveValue
        return (
            <div className={that.props.className || ''}>
                {that.props.canEdit && (
                    <SaveToolbar
                        dirty={isDirty}
                        disabled={that.props.loading || that.props.saving || !isDirty}
                        saving={that.props.saving}
                        onSave={that.onSave}
                        onDiscard={that.discard}
                    />
                )}
                {that.props.actions && (
                    <div className="site-admin-configuration-page__action-groups">
                        <div className="site-admin-configuration-page__actions">
                            {that.props.actions.map(({ id, label }) => (
                                <button
                                    key={id}
                                    className="btn btn-secondary btn-sm site-admin-configuration-page__action"
                                    onClick={() => that.runAction(id, that.configEditor)}
                                    type="button"
                                >
                                    {label}
                                </button>
                            ))}
                        </div>
                    </div>
                )}
                <React.Suspense fallback={<LoadingSpinner className="icon-inline mt-2" />}>
                    <MonacoSettingsEditor
                        {...that.props}
                        onDidSave={that.onSave}
                        onChange={that.onChange}
                        value={effectiveValue}
                        monacoRef={that.monacoRef}
                    />
                </React.Suspense>
            </div>
        )
    }

    private onSave = (): void => {
        const value = that.effectiveValue
        if (that.props.onSave) {
            that.props.onSave(value)
        }
    }

    private onChange = (newValue: string): void => {
        that.setState({ value: newValue }, () => {
            if (that.props.onChange) {
                that.props.onChange(newValue)
            }
            if (that.props.onDirtyChange) {
                that.props.onDirtyChange(that.isDirty)
            }
        })
    }

    private discard = (): void => {
        if (
            that.state.value === undefined ||
            that.props.value === that.state.value ||
            window.confirm('Discard edits?')
        ) {
            that.setState({ value: undefined })
        }
    }

    private monacoRef = (monacoValue: typeof _monaco | null): void => {
        that.monaco = monacoValue
        if (that.monaco && MonacoSettingsEditor) {
            that.subscriptions.add(
                disposableToFn(
                    that.monaco.editor.onDidCreateEditor(editor => {
                        that.configEditor = editor
                    })
                )
            )
            that.subscriptions.add(
                disposableToFn(
                    that.monaco.editor.onDidCreateModel(async model => {
                        // This function can only be called if the lazy MonacoSettingsEditor component was loaded,
                        // so that import call will not incur another load.
                        const { MonacoSettingsEditor } = await import('./MonacoSettingsEditor')

                        if (
                            that.configEditor &&
                            MonacoSettingsEditor.isStandaloneCodeEditor(that.configEditor) &&
                            that.props.actions
                        ) {
                            for (const { id, label, run } of that.props.actions) {
                                MonacoSettingsEditor.addEditorAction(that.configEditor, model, label, id, run)
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
            action.run().then(
                () => undefined,
                (err: any) => console.error(err)
            )
        } else {
            alert('Wait for editor to load before running action.')
        }
    }
}
