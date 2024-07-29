import React from 'react'

import type { JSONSchemaType } from 'ajv'
import classNames from 'classnames'
import { cloneDeep } from 'lodash'
import type * as monaco from 'monaco-editor/esm/vs/editor/editor.api'

import 'monaco-yaml'

import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter, map, startWith, switchMap } from 'rxjs/operators'

import { dataOrThrowErrors } from '@sourcegraph/http-client'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'

import batchSpecSchemaJSON from '../../../../../../../../schema/batch_spec.schema.json'
import { requestGraphQL } from '../../../../../backend/graphql'
import type {
    BatchSpecExecutionAvailableSecretKeysResult,
    BatchSpecExecutionAvailableSecretKeysVariables,
    Scalars,
} from '../../../../../graphql-operations'

import { BATCH_SPEC_EXECUTION_AVAILABLE_SECRET_KEYS } from './backend'

import styles from './MonacoBatchSpecEditor.module.scss'

/**
 * Minimal shape of a JSON Schema. These values are treated as opaque, so more specific types are
 * not needed.
 */
interface JSONSchema {
    $id: string
}

export interface Props {
    className?: string
    batchChangeNamespace?: { id: Scalars['ID']; __typename: 'User' | 'Org' }
    batchChangeName: string
    value: string | undefined
    isLightTheme: boolean
    onChange?: (newValue: string) => void
    readOnly?: boolean | undefined
    autoFocus?: boolean
}

interface State {
    availableSecrets?: string[]
}

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

        // We fetch the execution secrets available in this namespace to provide
        // autocomplete for env fields. The secrets are namespace-specific.
        this.subscriptions.add(
            componentUpdates
                .pipe(
                    filter(props => !props.readOnly && props.batchChangeNamespace !== undefined),
                    distinctUntilChanged((prev, next) => prev.batchChangeNamespace === next.batchChangeNamespace),
                    switchMap(props =>
                        requestGraphQL<
                            BatchSpecExecutionAvailableSecretKeysResult,
                            BatchSpecExecutionAvailableSecretKeysVariables
                        >(BATCH_SPEC_EXECUTION_AVAILABLE_SECRET_KEYS, {
                            namespace: props.batchChangeNamespace!.id,
                        })
                    ),
                    map(dataOrThrowErrors)
                )
                .subscribe(data => {
                    if (data.node && (data.node.__typename === 'User' || data.node.__typename === 'Org')) {
                        this.setState({ availableSecrets: data.node.executorSecrets.nodes.map(node => node.key) })
                        if (this.monaco) {
                            setDiagnosticsOptions(
                                this.monaco,
                                this.props.batchChangeName,
                                data.node.executorSecrets.nodes.map(node => node.key)
                            )
                        }
                    }
                })
        )

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
                autoFocus={this.props.autoFocus}
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

        setDiagnosticsOptions(monaco, this.props.batchChangeName, this.state.availableSecrets)

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

function setDiagnosticsOptions(editor: typeof monaco, batchChangeName: string, availableSecrets?: string[]): void {
    const schema = cloneDeep(batchSpecSchemaJSON)

    // Temporarily remove the mount field from the schema, so it does not show up in the auto-suggestion.

    // @ts-ignore
    delete schema.properties.steps.items.properties.mount

    if (availableSecrets !== undefined) {
        // Rewrite the JSON schema so that the env field has proper auto completion and
        // warns if any secrets referenced don't exist.
        ;(schema.properties.steps.items.properties.env.oneOf[2].items!.oneOf[0] as JSONSchemaType<'string'>).examples =
            availableSecrets
        ;(schema.properties.steps.items.properties.env.oneOf[2].items!.oneOf[0] as JSONSchemaType<'string'>).enum =
            availableSecrets
    }

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
