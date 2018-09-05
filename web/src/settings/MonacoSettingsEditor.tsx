import * as jsonc from '@sqs/jsonc-parser'
import * as monaco from 'monaco-editor'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'
import contributionSchema from 'sourcegraph/lib/protocol/contribution.schema.json'
import { BuiltinTheme, MonacoEditor } from '../components/MonacoEditor'
import jsonSchemaMetaSchema from '../schema/json-schema.schema.json'
import settingsSchema from '../schema/settings.schema.json'
import { eventLogger } from '../tracking/eventLogger'

const isLightThemeToMonacoTheme = (isLightTheme: boolean): BuiltinTheme => (isLightTheme ? 'vs' : 'sourcegraph-dark')

export interface Props {
    id?: string
    value: string | undefined
    onChange?: (newValue: string) => void
    readOnly?: boolean | undefined
    height?: number

    /**
     * The JSON Schema that describes the document.
     */
    jsonSchema: { $id: string }

    monacoRef?: (monacoValue: typeof monaco | null) => void
    isLightTheme: boolean

    /**
     * Called when the user presses the key binding for "save" (Ctrl+S/Cmd+S).
     */
    onDidSave?: () => void
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
        this.subscriptions.add(
            componentUpdates
                .pipe(
                    map(props => props.isLightTheme),
                    distinctUntilChanged(),
                    map(isLightThemeToMonacoTheme)
                )
                .subscribe(monacoTheme => {
                    if (this.monaco) {
                        this.monaco.editor.setTheme(monacoTheme)
                    }
                })
        )
    }

    public componentWillReceiveProps(newProps: Props): void {
        this.componentUpdates.next(newProps)
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
                id={this.props.id}
                language="json"
                height={this.props.height || 400}
                theme={isLightThemeToMonacoTheme(this.props.isLightTheme)}
                value={this.props.value}
                editorWillMount={this.editorWillMount}
                options={{
                    lineNumbers: 'off',
                    automaticLayout: true,
                    minimap: { enabled: false },
                    formatOnType: true,
                    formatOnPaste: true,
                    autoIndent: true,
                    renderIndentGuides: false,
                    glyphMargin: false,
                    folding: false,
                    renderLineHighlight: 'none',
                    scrollBeyondLastLine: false,
                    quickSuggestions: true,
                    quickSuggestionsDelay: 200,
                    readOnly: this.props.readOnly,
                    wordWrap: 'on',
                }}
            />
        )
    }

    private editorWillMount = (e: typeof monaco) => {
        this.monaco = e
        if (e) {
            this.onDidEditorMount()
        }
    }

    private onDidEditorMount(): void {
        const monaco = this.monaco!

        if (this.props.monacoRef) {
            this.props.monacoRef(monaco)
            this.subscriptions.add(() => {
                if (this.props.monacoRef) {
                    this.props.monacoRef(null)
                }
            })
        }

        monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
            validate: true,
            allowComments: true,
            schemas: [
                {
                    uri: this.props.jsonSchema.$id,
                    schema: this.props.jsonSchema,
                    fileMatch: ['*'],
                },

                // Include these schemas because they are referenced by other schemas.
                {
                    uri: 'http://json-schema.org/draft-07/schema',
                    schema: jsonSchemaMetaSchema,
                },
                {
                    uri: 'https://sourcegraph.com/v1/settings.schema.json#',
                    schema: settingsSchema,
                },
                {
                    // This is the literal relative URI used in extension.schema.json to refer to the contributions
                    // JSON Schema.
                    uri: './contribution.schema.json',
                    schema: contributionSchema,
                },
                {
                    // This is the absolute URI of the contributions JSON Schema.
                    uri: 'https://sourcegraph.com/v1/contribution.schema.json#',
                    schema: contributionSchema,
                },
            ],
        })

        monaco.editor.defineTheme('sourcegraph-dark', {
            base: 'vs-dark',
            inherit: true,
            colors: {
                'editor.background': '#0E121B',
                'editor.foreground': '#F2F4F8',
                'editorCursor.foreground': '#A2B0CD',
                'editor.selectionBackground': '#1C7CD650',
                'editor.selectionHighlightBackground': '#1C7CD625',
                'editor.inactiveSelectionBackground': '#1C7CD625',
            },
            rules: [],
        })

        this.disposables.push(monaco.editor.onDidCreateEditor(editor => this.onDidCreateEditor(editor)))
        this.disposables.push(monaco.editor.onDidCreateModel(model => this.onDidCreateModel(model)))
    }

    private onDidCreateEditor(editor: monaco.editor.ICodeEditor): void {
        this.editor = editor

        // Necessary to wrap in setTimeout or else _standaloneKeyBindingService won't be ready and the editor will
        // refuse to add the command because it's missing the keybinding service.
        setTimeout(() => {
            if (isStandaloneCodeEditor(editor)) {
                editor.addCommand(
                    monaco.KeyMod.CtrlCmd | monaco.KeyCode.KEY_S,
                    () => {
                        if (this.props.onDidSave) {
                            this.props.onDidSave()
                        }
                    },
                    ''
                )
            }
        })
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

export function isStandaloneCodeEditor(
    editor: monaco.editor.ICodeEditor
): editor is monaco.editor.IStandaloneCodeEditor {
    return editor.getEditorType() === monaco.editor.EditorType.ICodeEditor
}

export function toMonacoEdits(
    model: monaco.editor.IModel,
    edits: jsonc.Edit[]
): monaco.editor.IIdentifiedSingleEditOperation[] {
    return edits.map(
        (edit, i) =>
            ({
                identifier: { major: model.getVersionId(), minor: i },
                range: monaco.Range.fromPositions(
                    model.getPositionAt(edit.offset),
                    model.getPositionAt(edit.offset + edit.length)
                ),
                forceMoveMarkers: true,
                text: edit.content,
            } as monaco.editor.IIdentifiedSingleEditOperation)
    )
}

/**
 * A helper function that modifies site configuration to configure specific
 * common things, such as syncing GitHub repositories.
 */
export type ConfigInsertionFunction = (
    configJSON: string
) => {
    /** The edits to make to the input configuration to insert the new configuration. */
    edits: jsonc.Edit[]

    /** Select text in inserted JSON. */
    selectText?: string

    /**
     * If set, the selection is an empty selection that begins at the left-hand match of selectText plus this
     * offset. For example, if selectText is "foo" and cursorOffset is 2, then the final selection will be a cursor
     * "|" positioned as "fo|o".
     */
    cursorOffset?: number
}

export function addEditorAction(
    inputEditor: monaco.editor.IStandaloneCodeEditor,
    model: monaco.editor.IModel,
    label: string,
    id: string,
    run: ConfigInsertionFunction
): void {
    inputEditor.addAction({
        label,
        id,
        run: editor => {
            eventLogger.log('SiteConfigurationActionExecuted', { id })
            editor.focus()
            editor.pushUndoStop()
            const { edits, selectText, cursorOffset } = run(editor.getValue())
            const monacoEdits = toMonacoEdits(model, edits)
            let selection: monaco.Selection | undefined
            if (typeof selectText === 'string') {
                const afterText = jsonc.applyEdits(editor.getValue(), edits)
                let offset = afterText.slice(edits[0].offset).indexOf(selectText)
                if (offset !== -1) {
                    offset += edits[0].offset
                    if (typeof cursorOffset === 'number') {
                        selection = monaco.Selection.fromPositions(
                            getPositionAt(afterText, offset + cursorOffset),
                            getPositionAt(afterText, offset + cursorOffset)
                        )
                    } else {
                        selection = monaco.Selection.fromPositions(
                            getPositionAt(afterText, offset),
                            getPositionAt(afterText, offset + selectText.length)
                        )
                    }
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

function getPositionAt(text: string, offset: number): monaco.IPosition {
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
