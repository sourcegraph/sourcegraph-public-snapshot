import * as React from 'react'
import MonacoEditor from 'react-monaco-editor'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'

interface Props {
    className: string
    value: string | undefined
    onChange: (newValue: string) => void
    readOnly: boolean
}

/**
 * A WIP settings editor using the Monaco editor.
 *
 * TODO(sqs):
 *
 * - Decide on how theming should propagate
 *   (https://github.com/sourcegraph/sourcegraph/pull/8543) and propagate it
 *   here, so that the theme updates upon switching.
 * - Use a real, comprehensive JSON schema for our config.
 * - Examine impact on bundle size.
 */
export class MonacoSettingsEditor extends React.PureComponent<Props> {
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
                height={400}
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

    private monacoTheme(isLightTheme = window.localStorage.getItem('light-theme') === 'true'): string {
        // TODO(sqs): the theme is not updated after switching until you reload the page
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

        monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
            validate: true,
            allowComments: true,
            schemas: [
                {
                    fileMatch: ['*'],
                    uri: 'http://example.com/sourcegraph-config-schema.json',
                    schema: {
                        type: 'object',
                        properties: {
                            'search.savedQueries': {
                                description: 'Saved search queries',
                                type: 'array',
                                items: {
                                    type: 'object',
                                    properties: {
                                        key: {
                                            type: 'string',
                                            description: 'Unique key for this query in this file',
                                        },
                                        description: {
                                            type: 'string',
                                            description: 'Description of this saved query',
                                        },
                                        query: {
                                            type: 'string',
                                            description: 'Query string',
                                        },
                                        scopeQuery: {
                                            type: 'string',
                                            description: 'Scope query string',
                                        },
                                    },
                                    required: ['key', 'description', 'query'],
                                },
                            },
                        },
                    },
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

        this.disposables.push(monaco.editor.onDidCreateEditor(editor => (this.editor = editor)))
        this.disposables.push(monaco.editor.onDidCreateModel(model => this.onDidCreateModel(model)))
    }

    private onDidCreateModel(model: monaco.editor.IModel): void {
        this.disposables.push(
            model.onDidChangeContent(() => {
                this.props.onChange(model.getValue())
            })
        )
    }
}
