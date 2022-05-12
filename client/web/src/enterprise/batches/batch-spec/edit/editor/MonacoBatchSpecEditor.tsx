import React from 'react'

import classNames from 'classnames'
import { cloneDeep } from 'lodash'
// eslint-disable-next-line import/order
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-yaml'

import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'

import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import batchSpecSchemaJSON from '../../../../../../../../schema/batch_spec.schema.json'

import styles from './MonacoBatchSpecEditor.module.scss'

/**
 * Minimal shape of a JSON Schema. These values are treated as opaque, so more specific types are
 * not needed.
 */
interface JSONSchema {
    $id: string
}

export interface Props extends ThemeProps {
    className?: string
    batchChangeName: string
    value: string | undefined
    onChange?: (newValue: string) => void
    readOnly?: boolean | undefined
}

interface State {}

/**
 * Editor for Batch specs using Monaco editor.
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
                className={classNames(styles.editor, this.props.className)}
                language="yaml"
                height="auto"
                isLightTheme={this.props.isLightTheme}
                value={this.props.value}
                editorWillMount={this.editorWillMount}
                options={{
                    lineNumbers: 'on',
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

        setDiagnosticsOptions(monaco, this.props.batchChangeName)

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

function setDiagnosticsOptions(editor: typeof monaco, batchChangeName: string): void {
    const schema = cloneDeep(batchSpecSchemaJSON)
    // We don't allow env forwarding in src-cli so we remove it from the schema
    // so that monaco can show the error inline.
    schema.properties.steps.items.properties.env.oneOf[2].items!.oneOf = schema.properties.steps.items.properties.env.oneOf[2].items!.oneOf.filter(
        type => type.type !== 'string'
    )

    // Enforce the exact name match. The user must use the settings UI to change the name.
    schema.properties.name.pattern = `^${batchChangeName}$`

    editor.languages.yaml.yamlDefaults.setDiagnosticsOptions({
        validate: true,
        isKubernetes: false,
        format: true,
        completion: true,
        hover: true,
        schemas: [
            {
                uri: 'file:///root',
                schema: schema as JSONSchema,
                fileMatch: ['*'],
            },
        ],
    })
    // TODO: Register rename provider for output variables.
    // TODO: Register document highlight provider for output variables.
}
