import * as monaco from 'monaco-editor'
import * as React from 'react'

export type BuiltinTheme = monaco.editor.BuiltinTheme | 'sourcegraph-dark'

interface Props {
    /** The contents of the document. */
    value?: string

    /** The language of the document. */
    language?: string

    /** The DOM element ID to use when rendering the component. Use for a11y, not DOM manipulation. */
    id?: string

    /** The height (in px) of the Monaco editor. */
    height: number

    /** The color theme for the editor. */
    theme: BuiltinTheme

    /** Called when the editor has mounted. */
    editorWillMount: (editor: typeof monaco) => void

    /** Options for the editor. */
    options: monaco.editor.IEditorOptions
}

interface State {}

export class MonacoEditor extends React.PureComponent<Props, State> {
    private editor: monaco.editor.ICodeEditor | undefined

    private setRef = (e: HTMLElement | null): void => {
        if (!e) {
            return
        }

        this.props.editorWillMount(monaco)
        this.editor = monaco.editor.create(e, {
            value: this.props.value,
            language: this.props.language,
            theme: this.props.theme,
            ...this.props.options,
        })
    }

    public componentWillReceiveProps(nextProps: Props): void {
        if (this.props.value !== nextProps.value) {
            if (this.editor && this.editor.getValue() !== nextProps.value) {
                this.editor.setValue(nextProps.value || '')
            }
        }
    }

    public componentWillUnmount(): void {
        if (this.editor) {
            this.editor.dispose()

            // HACK: Clean up ARIA container that Monaco apparently forgets to remove.
            // tslint:disable-next-line:ban
            document.querySelectorAll('.monaco-aria-container').forEach(e => e.remove())
        }
    }

    public render(): JSX.Element | null {
        return (
            <div
                // tslint:disable-next-line:jsx-ban-props
                style={{ height: `${this.props.height}px`, position: 'relative' }}
                ref={this.setRef}
                id={this.props.id}
                className="monaco-editor-container"
            />
        )
    }
}

window.MonacoEnvironment = {
    getWorkerUrl(moduleId: string, label: string): string {
        if (label === 'json') {
            return window.context.assetsRoot + '/scripts/json.worker.bundle.js'
        }
        return window.context.assetsRoot + '/scripts/editor.worker.bundle.js'
    },
}
