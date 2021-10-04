import classNames from 'classnames'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
// import 'monaco-yaml'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'

import batchSpecSchemaJSON from '../../../../../../../schema/batch_spec.schema.json'
import jsonSchemaMetaSchema from '../../../../../../../schema/json-schema-draft-07.schema.json'
import settingsSchema from '../../../../../../../schema/settings.schema.json'
import { MonacoEditor } from '../../../../components/MonacoEditor'

/**
 * Minimal shape of a JSON Schema. These values are treated as opaque, so more specific types are
 * not needed.
 */
interface JSONSchema {
    $id: string
}

export interface Props extends ThemeProps {
    className?: string
    value: string | undefined
    onChange?: (newValue: string) => void
    readOnly?: boolean | undefined
    height?: number

    // language?: string

    /**
     * JSON Schema of the document.
     */
    // jsonSchema?: JSONSchema

    // monacoRef?: (monacoValue: typeof monaco | null) => void

    /**
     * Called when the user presses the key binding for "save" (Ctrl+S/Cmd+S).
     */
    // onDidSave?: () => void
}

interface State {}

/**
 * A JSON settings editor using the Monaco editor.
 */
export class MonacoSettingsEditor extends React.PureComponent<Props, State> {
    public state: State = {}

    private monaco: typeof monaco | null = null
    private editor: monaco.editor.ICodeEditor | null = null

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private disposables: monaco.IDisposable[] = []

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        this.subscriptions.add(
            componentUpdates
                .pipe(
                    map(props => props.readOnly),
                    distinctUntilChanged()
                )
                .subscribe(readOnly => {
                    if (this.editor) {
                        this.editor.updateOptions({ readOnly })
                    }
                })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        for (const disposable of this.disposables) {
            disposable.dispose()
        }
        this.monaco = null
        this.editor = null
    }

    public render(): JSX.Element | null {
        return (
            <MonacoEditor
                className={classNames('monaco-settings-editor', this.props.className)}
                language="yaml"
                height={this.props.height || 400}
                isLightTheme={this.props.isLightTheme}
                value={this.props.value}
                editorWillMount={this.editorWillMount}
                options={{
                    lineNumbers: 'off',
                    automaticLayout: true,
                    minimap: { enabled: false },
                    formatOnType: true,
                    formatOnPaste: true,
                    renderIndentGuides: false,
                    glyphMargin: false,
                    folding: false,
                    renderLineHighlight: 'none',
                    scrollBeyondLastLine: false,
                    quickSuggestions: true,
                    quickSuggestionsDelay: 200,
                    wordBasedSuggestions: false,
                    readOnly: this.props.readOnly,
                    wordWrap: 'on',
                }}
            />
        )
    }

    private editorWillMount = (editor: typeof monaco): void => {
        this.monaco = editor
        if (editor) {
            this.onDidEditorMount()
        }
    }

    private onDidEditorMount(): void {
        const monaco = this.monaco!

        setDiagnosticsOptions(monaco)

        // Only listen to 1 event each to avoid receiving events from other Monaco editors on the
        // same page (if there are multiple).
        const editorDisposable = monaco.editor.onDidCreateEditor(editor => {
            this.onDidCreateEditor(editor)
            editorDisposable.dispose()
        })
        this.disposables.push(editorDisposable)
        const modelDisposable = monaco.editor.onDidCreateModel(model => {
            this.onDidCreateModel(model)
            modelDisposable.dispose()
        })
        this.disposables.push(modelDisposable)
    }

    private onDidCreateEditor(editor: monaco.editor.ICodeEditor): void {
        this.editor = editor
    }

    private onDidCreateModel(model: monaco.editor.IModel): void {
        this.disposables.push(
            model.onDidChangeContent(() => {
                if (this.props.onChange) {
                    this.props.onChange(model.getValue())
                }
            })
        )
    }
}

function setDiagnosticsOptions(editor: typeof monaco): void {
    editor.languages.yaml.yamlDefaults.setDiagnosticsOptions({
        validate: true,
        isKubernetes: false,
        format: true,
        completion: true,
        hover: true,
        schemas: [
            {
                uri: 'file:///root',
                schema: batchSpecSchemaJSON,
                fileMatch: ['*'],
            },
            // Include these schemas because they are referenced by other schemas.
            {
                uri: 'http://json-schema.org/draft-07/schema',
                schema: jsonSchemaMetaSchema as JSONSchema,
            },
            {
                uri: 'settings.schema.json#',
                schema: settingsSchema,
            },
            {
                uri: 'settings.schema.json',
                schema: settingsSchema,
            },
        ],
    })
    editor.languages.registerRenameProvider('yaml', {
        provideRenameEdits: (model, position, newName, token) => {
            if (
                (position.lineNumber === 7 && position.column >= 7 && position.column <= 11) ||
                (position.lineNumber === 9 && position.column >= 27 && position.column <= 31)
            ) {
                return {
                    edits: [
                        {
                            edit: {
                                range: { startLineNumber: 7, startColumn: 7, endLineNumber: 7, endColumn: 11 },
                                text: newName,
                            },
                            modelVersionId: model.getVersionId(),
                            resource: model.uri,
                        },
                        {
                            edit: {
                                range: { startLineNumber: 9, startColumn: 27, endLineNumber: 9, endColumn: 31 },
                                text: newName,
                            },
                            modelVersionId: model.getVersionId(),
                            resource: model.uri,
                        },
                    ],
                }
            }
            return null
        },
    })
    // Hover contribution. This highlights all uses of a variable when the cursor is over it.
    editor.languages.registerDocumentHighlightProvider('yaml', {
        provideDocumentHighlights: (model, position, token) => {
            if (
                (position.lineNumber === 7 && position.column >= 7 && position.column <= 11) ||
                (position.lineNumber === 9 && position.column >= 27 && position.column <= 31)
            ) {
                return [
                    {
                        range: {
                            startLineNumber: 7,
                            startColumn: 7,
                            endLineNumber: 7,
                            endColumn: 11,
                        },
                        kind: 2,
                    },
                    {
                        range: {
                            startLineNumber: 9,
                            startColumn: 27,
                            endLineNumber: 9,
                            endColumn: 31,
                        },
                        kind: 1,
                    },
                ]
            }
            return null
        },
    })
}
