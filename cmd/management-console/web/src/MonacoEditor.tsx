// Regular imports
import * as _monaco from 'monaco-editor'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, startWith } from 'rxjs/operators'
import * as criticalSchema from '../../../../schema/critical.schema.json'
import './MonacoEditor.scss'

// Monaco imports. These are manually specified due to Parcel / ESM (I think).
// You can find a full list of possible imports / editor features here:
//
// https://github.com/Microsoft/monaco-editor-samples/blob/master/browser-esm-parcel/src/index.js#L2-L91
//
import 'monaco-editor/esm/vs/editor/browser/controller/coreCommands'
import 'monaco-editor/esm/vs/editor/browser/widget/codeEditorWidget'
import 'monaco-editor/esm/vs/editor/contrib/bracketMatching/bracketMatching'
import 'monaco-editor/esm/vs/editor/contrib/caretOperations/caretOperations'
import 'monaco-editor/esm/vs/editor/contrib/caretOperations/transpose'
import 'monaco-editor/esm/vs/editor/contrib/clipboard/clipboard'
import 'monaco-editor/esm/vs/editor/contrib/codelens/codelensController'
import 'monaco-editor/esm/vs/editor/contrib/colorPicker/colorDetector'
import 'monaco-editor/esm/vs/editor/contrib/comment/comment'
import 'monaco-editor/esm/vs/editor/contrib/contextmenu/contextmenu'
import 'monaco-editor/esm/vs/editor/contrib/cursorUndo/cursorUndo'
import 'monaco-editor/esm/vs/editor/contrib/dnd/dnd'
import 'monaco-editor/esm/vs/editor/contrib/find/findController'
import 'monaco-editor/esm/vs/editor/contrib/folding/folding'
import 'monaco-editor/esm/vs/editor/contrib/format/formatActions'
import 'monaco-editor/esm/vs/editor/contrib/gotoError/gotoError'
import 'monaco-editor/esm/vs/editor/contrib/hover/hover'
import 'monaco-editor/esm/vs/editor/contrib/inPlaceReplace/inPlaceReplace'
import 'monaco-editor/esm/vs/editor/contrib/linesOperations/linesOperations'
import 'monaco-editor/esm/vs/editor/contrib/links/links'
import 'monaco-editor/esm/vs/editor/contrib/multicursor/multicursor'
import 'monaco-editor/esm/vs/editor/contrib/parameterHints/parameterHints'
import 'monaco-editor/esm/vs/editor/contrib/referenceSearch/referenceSearch'
import 'monaco-editor/esm/vs/editor/contrib/rename/rename'
import 'monaco-editor/esm/vs/editor/contrib/smartSelect/smartSelect'
import 'monaco-editor/esm/vs/editor/contrib/snippet/snippetController2'
import 'monaco-editor/esm/vs/editor/contrib/suggest/suggestController'
import 'monaco-editor/esm/vs/editor/contrib/wordHighlighter/wordHighlighter'
import 'monaco-editor/esm/vs/editor/contrib/wordOperations/wordOperations'
import * as monaco from 'monaco-editor/esm/vs/editor/editor.api'
import 'monaco-editor/esm/vs/editor/standalone/browser/inspectTokens/inspectTokens'

import 'monaco-editor/esm/vs/language/json/monaco.contribution'

interface Props {
    /**
     * The content the editor should display.
     */
    content: string

    /**
     * The language of the content (e.g. "json").
     */
    language: string

    /**
     * Called when the user changes the content of the editor.
     *
     * @param content the literal content of the editor
     */
    onDidContentChange(content: string): void

    /**
     * Called when the user presses the key binding for "save" (Ctrl+S/Cmd+S).
     */
    onDidSave: () => void

    /** Called when the editor will mount. */
    editorWillMount: (editor: monaco.editor.IStandaloneCodeEditor, model: monaco.editor.IModel) => void
}

export class MonacoEditor extends React.Component<Props, {}> {
    private ref: HTMLElement | null = null
    private editor: monaco.editor.IStandaloneCodeEditor | null = null
    private model: monaco.editor.IModel | null = null

    private componentUpdates = new Subject<Props>()
    private subscriptions = new Subscription()
    private disposables: monaco.IDisposable[] = []

    public componentDidMount(): void {
        const componentUpdates = this.componentUpdates.pipe(startWith(this.props))

        // TODO(slimsag): I do not understand why this cast is neccessary, and there must be a good reason
        monaco.editor.onDidCreateEditor(editor => this.onDidCreateEditor(editor as monaco.editor.IStandaloneCodeEditor))
        monaco.editor.onDidCreateModel(model => this.onDidCreateModel(model))

        this.subscriptions.add(
            componentUpdates
                .pipe(
                    map(props => [props.content, props.language]),
                    distinctUntilChanged()
                )
                .subscribe(([content, language]) => {
                    if (this.model) {
                        this.model.setValue(content)
                        monaco.editor.setModelLanguage(this.model, language)
                    }
                })
        )

        const modelUri = monaco.Uri.parse('a://b/foo.json') // a made up unique URI for our model
        const model = monaco.editor.createModel('', 'json', modelUri)

        monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
            allowComments: true,
            validate: true,
            schemas: [
                {
                    uri: 'https://fake-schema.org/critical-schema.json',
                    fileMatch: [modelUri.toString()], // associate with our model
                    schema: criticalSchema,
                },
            ],
        })

        // Create the actual Monaco editor.
        const editor = monaco.editor.create(this.ref!, {
            lineNumbers: 'on',
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
            wordBasedSuggestions: false,
            wordWrap: 'on',
            theme: 'vs-dark',
            model,
        })

        this.props.editorWillMount(editor, model)

        // Register theme for the editor.
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
        monaco.editor.setTheme('sourcegraph-dark')
    }

    public componentWillUnmount(): void {
        // TODO(slimsag): future: does this actually teardown Monaco properly?
        this.subscriptions.unsubscribe()
        for (const disposable of this.disposables) {
            disposable.dispose()
        }
        this.ref = null
        this.editor = null
    }

    private onDidCreateEditor = (editor: monaco.editor.IStandaloneCodeEditor): void => {
        this.editor = editor
    }

    private onDidCreateModel = (model: monaco.editor.IModel): void => {
        this.model = model

        // Necessary to wrap in setTimeout or else _standaloneKeyBindingService
        // won't be ready and the editor will refuse to add the command because
        // it's missing the keybinding service.
        setTimeout(() => {
            this.editor!.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KEY_S, () => this.props.onDidSave(), '')
        })

        this.model.setValue(this.props.content)
        monaco.editor.setModelLanguage(this.model, this.props.language)
        this.model.updateOptions({ tabSize: 2 })

        model.onDidChangeContent(e => {
            this.props.onDidContentChange(model.getValue())
        })
    }

    public render(): JSX.Element | null {
        return <div className="monaco-editor-container" ref={this.setRef} />
    }

    private setRef = (e: HTMLElement | null): void => {
        this.ref = e
    }
}

// TODO(slimsag): future: This code is correct, but I do not know how to get
// proper typings imported for this. Presumably I need to pull in some .d.ts file?
;(self as any).MonacoEnvironment = {
    getWorker(moduleId: any, label: string): Worker {
        if (label === 'json') {
            return new Worker('../node_modules/monaco-editor/esm/vs/language/json/json.worker.js')
        }
        if (label === 'css') {
            return new Worker('../node_modules/monaco-editor/esm/vs/language/css/css.worker.js')
        }
        if (label === 'html') {
            return new Worker('../node_modules/monaco-editor/esm/vs/language/html/html.worker.js')
        }
        if (label === 'typescript' || label === 'javascript') {
            return new Worker('../node_modules/monaco-editor/esm/vs/language/typescript/ts.worker.js')
        }
        return new Worker('../node_modules/monaco-editor/esm/vs/editor/editor.worker.js')
    },
}
