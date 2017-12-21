import * as React from 'react'
import MonacoEditor from 'react-monaco-editor'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import SettingsSchemaJSON from './settings.schema.json'
import SiteSchemaJSON from './site.schema.json'
import { colorTheme } from './theme'

interface Props {
    className: string
    value: string | undefined
    onChange?: (newValue: string) => void
    readOnly: boolean
    height?: number

    /**
     * The ID of the JSON Schema that describes the document.
     */
    jsonSchema: string
}

interface State {
    isLightTheme?: boolean
}

/**
 * A JSON settings editor using the Monaco editor.
 */
export class MonacoSettingsEditor extends React.PureComponent<Props, State> {
    public state: State = {}

    private monaco: typeof monaco | null
    private editor: monaco.editor.ICodeEditor

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private disposables: monaco.IDisposable[] = []

    constructor(props: Props) {
        super(props)

        this.subscriptions.add(
            this.componentUpdates
                .pipe(startWith(props), map(props => props.readOnly), distinctUntilChanged())
                .subscribe(readOnly => {
                    if (this.editor) {
                        this.editor.updateOptions({ readOnly })
                    }
                })
        )
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            colorTheme.subscribe(theme => {
                this.setState({ isLightTheme: theme === 'light' }, () => {
                    if (this.monaco) {
                        this.monaco.editor.setTheme(this.monacoTheme())
                    }
                })
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
    }

    public render(): JSX.Element | null {
        return (
            <MonacoEditor
                language="json"
                height={this.props.height || 400}
                theme={this.monacoTheme()}
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
                    quickSuggestionsDelay: 200,
                }}
                requireConfig={{ paths: { vs: '/.assets/scripts/vs' }, url: '/.assets/scripts/vs/loader.js' }}
            />
        )
    }

    private monacoTheme(isLightTheme = this.state.isLightTheme): string {
        return isLightTheme ? 'vs' : 'sourcegraph-dark'
    }

    private editorWillMount = (e: typeof monaco) => {
        this.monaco = e
        if (e) {
            this.onDidEditorMount()
        }
    }

    private onDidEditorMount(): void {
        const monaco = this.monaco!

        const schemas: { uri: string; schema: any }[] = [
            { uri: 'https://sourcegraph.com/v1/site.schema.json#', schema: SiteSchemaJSON },
            { uri: 'https://sourcegraph.com/v1/settings.schema.json#', schema: SettingsSchemaJSON },
        ]
        monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
            validate: true,
            allowComments: true,
            schemas: schemas.map(schema => ({
                ...schema,
                fileMatch: schema.uri === this.props.jsonSchema ? ['*'] : undefined,
            })),
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

        this.disposables.push(monaco.editor.onDidCreateEditor(editor => (this.editor = editor)))
        this.disposables.push(monaco.editor.onDidCreateModel(model => this.onDidCreateModel(model)))
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
