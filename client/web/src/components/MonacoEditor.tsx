import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import * as monaco from 'monaco-editor'
import * as React from 'react'
import { Subscription, Subject } from 'rxjs'
import { map, distinctUntilChanged } from 'rxjs/operators'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

const SOURCEGRAPH_LIGHT = 'sourcegraph-light'
const SOURCEGRAPH_DARK = 'sourcegraph-dark'

// 🚨 WARNING!!!
// Monaco does not support CSS variables/custom properties, all colors must be in hex codes.
// See https://github.com/microsoft/monaco-editor/issues/2427
// When updating these colors, always add the name of the color variable from CSS so
// we can look up uses later when updating color palettes.

const darkColors: monaco.editor.IColors = {
    background: '#181b26', // --color-bg-1
    'editor.background': '#181b26', // --color-bg-1
    'textLink.activeBackground': '#1d212f', // --color-bg-2
    'editor.foreground': '#dbe2f0', // --search-query-text-color
    'editorCursor.foreground': '#dbe2f0', // --search-query-text-color
    'editorSuggestWidget.background': '#181b26', // --color-bg-1
    'editorSuggestWidget.foreground': '#dbe2f0', // --search-query-text-color
    'editorSuggestWidget.highlightForeground': '#4393e7', // --search-filter-keyword-color
    'editorSuggestWidget.selectedBackground': '#1d212f', // --color-bg-2
    'list.hoverBackground': '#1d212f', // --color-bg-2
    'editorSuggestWidget.border': '#262b38', // --border-color
    'editorHoverWidget.background': '#181b26', // --color-bg-1
    'editorHoverWidget.foreground': '#dbe2f0', // --search-query-text-color
    'editorHoverWidget.border': '#262b38', // --border-color
    'editor.hoverHighlightBackground': '#1d212f', // --color-bg-2
}

const darkRules: monaco.editor.ITokenThemeRule[] = [
    // Sourcegraph base language tokens
    { token: 'identifier', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'field', foreground: '#4393e7' }, // --search-filter-keyword-color
    { token: 'keyword', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'openingParen', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'closingParen', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'comment', foreground: '#ffa94d' }, // --oc-orange-4
    // Sourcegraph decorated language tokens
    { token: 'metaRepoRevisionSeparator', foreground: '#4393e7' }, // --search-filter-keyword-color
    { token: 'metaContextPrefix', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'metaPredicateNameAccess', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'metaPredicateDot', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaPredicateParenthesis', foreground: '#ffa94d' }, // --oc-orange-4
    // Regexp pattern highlighting
    { token: 'metaRegexpDelimited', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaRegexpAssertion', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaRegexpLazyQuantifier', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaRegexpEscapedCharacter', foreground: '#ffa8a8' }, // --oc-red-3
    { token: 'metaRegexpCharacterSet', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'metaRegexpCharacterClass', foreground: '#da77f2' }, // --oc-grape-4
    { token: 'metaRegexpCharacterClassMember', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaRegexpCharacterClassRange', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#d68cf3' }, // --search-keyword-color
    { token: 'metaRegexpRangeQuantifier', foreground: '#3bc9db' }, // --oc-cyan-4
    { token: 'metaRegexpAlternative', foreground: '#3bc9db' }, // --oc-cyan-4
    // Structural pattern highlighting
    { token: 'metaStructuralHole', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaStructuralRegexpHole', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaStructuralVariable', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaStructuralRegexpSeparator', foreground: '#ffa94d' }, // --oc-orange-4
    // Revision highlighting
    { token: 'metaRevisionSeparator', foreground: '#ffa94d' }, // --oc-orange-4
    { token: 'metaRevisionIncludeGlobMarker', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaRevisionExcludeGlobMarker', foreground: '#ff6b6b' }, // --oc-red-5
    { token: 'metaRevisionCommitHash', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaRevisionLabel', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaRevisionReferencePath', foreground: '#dbe2f0' }, // --search-query-text-color
    { token: 'metaRevisionWildcard', foreground: '#3bc9db' }, // --oc-cyan-4
    // Path-like highlighting
    { token: 'metaPathSeparator', foreground: '#868e96' }, // --oc-gray-6
]

const lightColors: monaco.editor.IColors = {
    background: '#ffffff', // --color-bg-1
    'editor.background': '#ffffff', // --color-bg-1
    'editor.foreground': '#14171f', // --search-query-text-color
    'editorCursor.foreground': '#14171f', // --search-query-text-color
    'editorSuggestWidget.background': '#ffffff', // --color-bg-1
    'editorSuggestWidget.foreground': '#14171f', // --search-query-text-color
    'editorSuggestWidget.border': '#e6ebf2', // --border-color
    'editorSuggestWidget.highlightForeground': '#0b70db', // --search-filter-keyword-color
    'editorSuggestWidget.selectedBackground': '#e6ebf2', // --color-bg-2
    'list.hoverBackground': '#e6ebf2', // --color-bg-2
    'editorHoverWidget.background': '#ffffff', // --color-bg-1
    'editorHoverWidget.foreground': '#14171f', // --search-query-text-color
    'editorHoverWidget.border': '#e6ebf2', // --border-color
    'editor.hoverHighlightBackground': '#e6ebf2', // --color-bg-2
}

const lightRules: monaco.editor.ITokenThemeRule[] = [
    // Sourcegraph base language tokens
    { token: 'identifier', foreground: '#14171f' }, // --search-query-text-color
    { token: 'field', foreground: '#0b70db' }, // --search-filter-keyword-color
    { token: 'keyword', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'openingParen', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'closingParen', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'comment', foreground: '#d9480f' }, // --oc-orange-9
    // Sourcegraph decorated language tokens
    { token: 'metaRepoRevisionSeparator', foreground: '#0b70db' }, // --search-filter-keyword-color
    { token: 'metaContextPrefix', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'metaPredicateNameAccess', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'metaPredicateDot', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaPredicateParenthesis', foreground: '#d9480f' }, // --oc-orange-9
    // Regexp pattern highlighting
    { token: 'metaRegexpDelimited', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaRegexpAssertion', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaRegexpLazyQuantifier', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaRegexpEscapedCharacter', foreground: '#d9480f' }, // --oc-orange-9
    { token: 'metaRegexpCharacterSet', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'metaRegexpCharacterClass', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'metaRegexpCharacterClassMember', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaRegexpCharacterClassRange', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#a305e1' }, // --search-keyword-color
    { token: 'metaRegexpRangeQuantifier', foreground: '#1098ad' }, // --oc-cyan-7
    { token: 'metaRegexpAlternative', foreground: '#1098ad' }, // --oc-cyan-7
    // Structural pattern highlighting
    { token: 'metaStructuralHole', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaStructuralRegexpHole', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaStructuralVariable', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaStructuralRegexpSeparator', foreground: '#d9480f' }, // --oc-orange-9
    // Revision highlighting
    { token: 'metaRevisionSeparator', foreground: '#d9480f' }, // --oc-orange-9
    { token: 'metaRevisionIncludeGlobMarker', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaRevisionExcludeGlobMarker', foreground: '#c92a2a' }, // --oc-red-9
    { token: 'metaRevisionCommitHash', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaRevisionLabel', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaRevisionReferencePath', foreground: '#14171f' }, // --search-query-text-color
    { token: 'metaRevisionWildcard', foreground: '#1098ad' }, // --oc-cyan-7
    // Path-like highlighting
    { token: 'metaPathSeparator', foreground: '#868e96' }, // --oc-gray-6
]

monaco.editor.defineTheme(SOURCEGRAPH_DARK, {
    base: 'vs-dark',
    inherit: true,
    colors: darkColors,
    rules: darkRules,
})

monaco.editor.defineTheme(SOURCEGRAPH_LIGHT, {
    base: 'vs',
    inherit: true,
    colors: lightColors,
    rules: lightRules,
})

interface Props extends ThemeProps {
    /** The contents of the document. */
    value?: string

    /** The language of the document. */
    language?: string

    /** The DOM element ID to use when rendering the component. Use for a11y, not DOM manipulation. */
    id?: string

    /** The height (in px) of the Monaco editor. */
    height: number

    /** Called when the editor has mounted. */
    editorWillMount: (editor: typeof monaco) => void

    /** Called when a standalone code editor has been created with the given props */
    onEditorCreated?: (editor: monaco.editor.IStandaloneCodeEditor) => void

    /** Options for the editor. */
    options: monaco.editor.IStandaloneEditorConstructionOptions

    /** An optional className to add to the editor. */
    className?: string

    /** Whether to add a border to the Monaco editor. Default: true. */
    border?: boolean

    /** Keyboard shortcut to focus the Monaco editor. */
    keyboardShortcutForFocus?: KeyboardShortcut
}

interface State {}

export class MonacoEditor extends React.PureComponent<Props, State> {
    private subscriptions = new Subscription()

    private componentUpdates = new Subject<Props>()

    private editor: monaco.editor.ICodeEditor | undefined

    private setRef = (element: HTMLElement | null): void => {
        if (!element) {
            return
        }
        this.props.editorWillMount(monaco)
        const editor = monaco.editor.create(element, {
            value: this.props.value,
            language: this.props.language,
            theme: this.getTheme(this.props.isLightTheme),
            ...this.props.options,
        })
        if (this.props.onEditorCreated) {
            this.props.onEditorCreated(editor)
        }
        this.editor = editor
    }

    public componentDidUpdate(previousProps: Props): void {
        if (this.props.value !== previousProps.value && this.editor && this.editor.getValue() !== this.props.value) {
            this.editor.setValue(this.props.value || '')
        }
        this.componentUpdates.next(this.props)
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ isLightTheme }) => this.getTheme(isLightTheme)),
                    distinctUntilChanged()
                )
                .subscribe(theme => monaco.editor.setTheme(theme))
        )
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
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
            <>
                <div
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ height: `${this.props.height}px`, position: 'relative' }}
                    ref={this.setRef}
                    id={this.props.id}
                    className={classNames(this.props.className, this.props.border !== false && 'border')}
                />
                {this.props.keyboardShortcutForFocus?.keybindings.map((keybinding, index) => (
                    <Shortcut key={index} {...keybinding} onMatch={this.focusInput} />
                ))}
            </>
        )
    }

    private focusInput = (): void => {
        if (
            this.editor &&
            !!document.activeElement &&
            !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)
        ) {
            this.editor.focus()
        }
    }

    private getTheme = (isLightTheme: boolean): string => (isLightTheme ? SOURCEGRAPH_LIGHT : SOURCEGRAPH_DARK)
}

if (!window.MonacoEnvironment) {
    window.MonacoEnvironment = {
        getWorkerUrl(moduleId: string, label: string): string {
            if (label === 'json') {
                return window.context.assetsRoot + '/scripts/json.worker.bundle.js'
            }
            return window.context.assetsRoot + '/scripts/editor.worker.bundle.js'
        },
    }
}
