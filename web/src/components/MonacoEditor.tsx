import classNames from 'classnames'
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
    className?: string
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

    public componentDidUpdate(prevProps: Props): void {
        if (this.props.value !== prevProps.value && this.editor && this.editor.getValue() !== this.props.value) {
            this.editor.setValue(this.props.value || '')
        }
    }

    public componentWillUnmount(): void {
        if (this.editor) {
            this.editor.dispose()

            // HACK: Clean up ARIA container that Monaco apparently forgets to remove.
            for (const element of document.querySelectorAll('.monaco-aria-container')) {
                element.remove()
            }
        }
    }

    public render(): JSX.Element | null {
        return (
            <div
                // eslint-disable-next-line react/forbid-dom-props
                style={{ height: `${this.props.height}px`, position: 'relative' }}
                ref={this.setRef}
                id={this.props.id}
                className={classNames(this.props.className, 'border')}
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
