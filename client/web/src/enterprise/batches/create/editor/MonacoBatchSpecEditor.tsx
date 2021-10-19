import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-yaml'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { MonacoEditor } from '@sourcegraph/web/src/components/MonacoEditor'

import batchSpecSchemaJSON from '../../../../../../../schema/batch_spec.schema.json'
import jsonSchemaMetaSchema from '../../../../../../../schema/json-schema-draft-07.schema.json'

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
}

interface State {}

/**
 * A JSON settings editor using the Monaco editor.
 */
export class MonacoBatchSpecEditor extends React.PureComponent<Props, State> {
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
                className={this.props.className}
                language="yaml"
                height={this.props.height || 400}
                isLightTheme={this.props.isLightTheme}
                value={this.props.value}
                editorWillMount={this.editorWillMount}
                options={{
                    lineNumbers: 'on',
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
        ],
    })
    // TODO: Register rename provider for output variables.
    // TODO: Register document highlight provider for output variables.
}
