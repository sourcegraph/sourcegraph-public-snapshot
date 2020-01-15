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

        that.props.editorWillMount(monaco)
        that.editor = monaco.editor.create(e, {
            value: that.props.value,
            language: that.props.language,
            theme: that.props.theme,
            ...that.props.options,
        })
    }

    public componentDidUpdate(prevProps: Props): void {
        if (that.props.value !== prevProps.value && that.editor && that.editor.getValue() !== that.props.value) {
            that.editor.setValue(that.props.value || '')
        }
    }

    public componentWillUnmount(): void {
        if (that.editor) {
            that.editor.dispose()

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
                style={{ height: `${that.props.height}px`, position: 'relative' }}
                ref={that.setRef}
                id={that.props.id}
                className={classNames(that.props.className, 'border')}
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
