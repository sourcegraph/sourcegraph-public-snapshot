import { Shortcut } from '@slimsag/react-shortcuts'
import classNames from 'classnames'
import * as monaco from 'monaco-editor'
import * as React from 'react'
import { Subscription, Subject } from 'rxjs'
import { map, distinctUntilChanged } from 'rxjs/operators'

import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

const SOURCEGRAPH_LIGHT = 'sourcegraph-light'
const SOURCEGRAPH_LIGHT_REDESIGN = 'sourcegraph-light-redesign'

const SOURCEGRAPH_DARK = 'sourcegraph-dark'
const SOURCEGRAPH_DARK_REDESIGN = 'sourcegraph-dark-redesign'

const darkColors: monaco.editor.IColors = {
    background: '#0E121B',
    'editor.background': '#0E121B',
    'textLink.activeBackground': '#2a3a51',
    'editor.foreground': '#f2f4f8',
    'editorCursor.foreground': '#ffffff',
    'editorSuggestWidget.background': '#1c2736',
    'editorSuggestWidget.foreground': '#F2F4F8',
    'editorSuggestWidget.highlightForeground': '#569cd6',
    'editorSuggestWidget.selectedBackground': '#2a3a51',
    'list.hoverBackground': '#2a3a51',
    'editorSuggestWidget.border': '#2b3750',
    'editorHoverWidget.background': '#1c2736',
    'editorHoverWidget.foreground': '#F2F4F8',
    'editorHoverWidget.border': '#2b3750',
    'editor.hoverHighlightBackground': '#495057',
}

const darkRules: monaco.editor.ITokenThemeRule[] = [
    // Sourcegraph base language tokens
    { token: 'identifier', foreground: '#f2f4f8' },
    { token: 'field', foreground: '#569cd6' },
    { token: 'keyword', foreground: '#da77f2' },
    { token: 'openingParen', foreground: '#da77f2' },
    { token: 'closingParen', foreground: '#da77f2' },
    { token: 'comment', foreground: '#ffa94d' },
    // Sourcegraph decorated language tokens
    { token: 'metaRepoRevisionSeparator', foreground: '#569cd9' },
    { token: 'metaContextPrefix', foreground: '#da77f2' },
    { token: 'metaPredicateNameAccess', foreground: '#da77f2' },
    { token: 'metaPredicateDot', foreground: '#f2f4f8' },
    { token: 'metaPredicateParenthesis', foreground: '#f08d58' },
    // Regexp pattern highlighting
    { token: 'metaRegexpDelimited', foreground: '#ff6b6b' },
    { token: 'metaRegexpAssertion', foreground: '#ff6b6b' },
    { token: 'metaRegexpLazyQuantifier', foreground: '#ff6b6b' },
    { token: 'metaRegexpEscapedCharacter', foreground: '#ffa8a8' },
    { token: 'metaRegexpCharacterSet', foreground: '#da77f2' },
    { token: 'metaRegexpCharacterClass', foreground: '#da77f2' },
    { token: 'metaRegexpCharacterClassMember', foreground: '#f2f4f8' },
    { token: 'metaRegexpCharacterClassRange', foreground: '#f2f4f8' },
    { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#da77f2' },
    { token: 'metaRegexpRangeQuantifier', foreground: '#3bc9db' },
    { token: 'metaRegexpAlternative', foreground: '#3bc9db' },
    // Structural pattern highlighting
    { token: 'metaStructuralHole', foreground: '#ff6b6b' },
    { token: 'metaStructuralRegexpHole', foreground: '#ff6b6b' },
    { token: 'metaStructuralVariable', foreground: '#f2f4f8' },
    { token: 'metaStructuralRegexpSeparator', foreground: '#ffa94d' },
    // Revision highlighting
    { token: 'metaRevisionSeparator', foreground: '#ffa94d' },
    { token: 'metaRevisionIncludeGlobMarker', foreground: '#ff6b6b' },
    { token: 'metaRevisionExcludeGlobMarker', foreground: '#ff6b6b' },
    { token: 'metaRevisionCommitHash', foreground: '#f2f4f8' },
    { token: 'metaRevisionLabel', foreground: '#f2f4f8' },
    { token: 'metaRevisionReferencePath', foreground: '#f2f4f8' },
    { token: 'metaRevisionWildcard', foreground: '#3bc9db' },
    // Path-like highlighting
    { token: 'metaPathSeparator', foreground: '#868e96' },
]

const lightColors: monaco.editor.IColors = {
    background: '#ffffff',
    'editor.background': '#ffffff',
    'editor.foreground': '#2b3750',
    'editorCursor.foreground': '#2b3750',
    'editorSuggestWidget.background': '#ffffff',
    'editorSuggestWidget.foreground': '#2b3750',
    'editorSuggestWidget.border': '#cad2e2',
    'editorSuggestWidget.highlightForeground': '#268bd2',
    'editorSuggestWidget.selectedBackground': '#f2f4f8',
    'list.hoverBackground': '#f2f4f8',
    'editorHoverWidget.background': '#ffffff',
    'editorHoverWidget.foreground': '#2b3750',
    'editorHoverWidget.border': '#cad2e2',
    'editor.hoverHighlightBackground': '#dee2e6',
}

const lightRules: monaco.editor.ITokenThemeRule[] = [
    // Sourcegraph base language tokens
    { token: 'identifier', foreground: '#2b3750' },
    { token: 'field', foreground: '#268bd2' },
    { token: 'keyword', foreground: '#ae3ec9' },
    { token: 'openingParen', foreground: '#ae3ec9' },
    { token: 'closingParen', foreground: '#ae3ec9' },
    { token: 'comment', foreground: '#d9480f' },
    // Sourcegraph decorated language tokens
    { token: 'metaRepoRevisionSeparator', foreground: '#268bd2' },
    { token: 'metaContextPrefix', foreground: '#ae3ec9' },
    { token: 'metaPredicateNameAccess', foreground: '#ae3ec9' },
    { token: 'metaPredicateDot', foreground: '#2b3750' },
    { token: 'metaPredicateParenthesis', foreground: '#d6550f' },
    // Regexp pattern highlighting
    { token: 'metaRegexpDelimited', foreground: '#c92a2a' },
    { token: 'metaRegexpAssertion', foreground: '#c92a2a' },
    { token: 'metaRegexpLazyQuantifier', foreground: '#c92a2a' },
    { token: 'metaRegexpEscapedCharacter', foreground: '#af5200' },
    { token: 'metaRegexpCharacterSet', foreground: '#ae3ec9' },
    { token: 'metaRegexpCharacterClass', foreground: '#ae3ec9' },
    { token: 'metaRegexpCharacterClassMember', foreground: '#2b3750' },
    { token: 'metaRegexpCharacterClassRange', foreground: '#2b3750' },
    { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#ae3ec9' },
    { token: 'metaRegexpRangeQuantifier', foreground: '#1098ad' },
    { token: 'metaRegexpAlternative', foreground: '#1098ad' },
    // Structural pattern highlighting
    { token: 'metaStructuralHole', foreground: '#c92a2a' },
    { token: 'metaStructuralRegexpHole', foreground: '#c92a2a' },
    { token: 'metaStructuralVariable', foreground: '#2b3750' },
    { token: 'metaStructuralRegexpSeparator', foreground: '#d9480f' },
    // Revision highlighting
    { token: 'metaRevisionSeparator', foreground: '#d9480f' },
    { token: 'metaRevisionIncludeGlobMarker', foreground: '#c92a2a' },
    { token: 'metaRevisionExcludeGlobMarker', foreground: '#c92a2a' },
    { token: 'metaRevisionCommitHash', foreground: '#2b3750' },
    { token: 'metaRevisionLabel', foreground: '#2b3750' },
    { token: 'metaRevisionReferencePath', foreground: '#2b3750' },
    { token: 'metaRevisionWildcard', foreground: '#1098ad' },
    // Path-like highlighting
    { token: 'metaPathSeparator', foreground: '#868e96' },
]

monaco.editor.defineTheme(SOURCEGRAPH_DARK_REDESIGN, {
    base: 'vs-dark',
    inherit: true,
    colors: {
        ...darkColors,
        background: '#181b26',
        'editor.background': '#181b26',
        'editorSuggestWidget.border': '#343a4d',
        'editorHoverWidget.border': '#343a4d',
    },
    rules: [
        ...darkRules,
        // Sourcegraph base language tokens
        { token: 'identifier', foreground: '#f9fafb' },
        { token: 'field', foreground: '#4393e7' },
        { token: 'keyword', foreground: '#d68cf3' },
        { token: 'openingParen', foreground: '#d68cf3' },
        { token: 'closingParen', foreground: '#d68cf3' },
        // Sourcegraph decorated language tokens
        { token: 'metaRepoRevisionSeparator', foreground: '#569cd9' },
        { token: 'metaContextPrefix', foreground: '#d68cf3' },
        { token: 'metaPredicateNameAccess', foreground: '#d68cf3' },
        { token: 'metaPredicateDot', foreground: '#f9fafb' },
        // Regexp pattern highlighting
        { token: 'metaRegexpCharacterSet', foreground: '#d68cf3' },
        { token: 'metaRegexpCharacterClass', foreground: '#d68cf3' },
        { token: 'metaRegexpCharacterClassMember', foreground: '#f9fafb' },
        { token: 'metaRegexpCharacterClassRange', foreground: '#f9fafb' },
        { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#d68cf3' },
        // Structural pattern highlighting
        { token: 'metaStructuralVariable', foreground: '#f9fafb' },
        // Revision highlighting
        { token: 'metaRevisionCommitHash', foreground: '#f9fafb' },
        { token: 'metaRevisionLabel', foreground: '#f9fafb' },
        { token: 'metaRevisionReferencePath', foreground: '#f9fafb' },
    ],
})

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

monaco.editor.defineTheme(SOURCEGRAPH_LIGHT_REDESIGN, {
    base: 'vs',
    inherit: true,
    colors: {
        ...lightColors,
        'editor.foreground': '#14171f',
        'editorCursor.foreground': '#14171f',
        'editorSuggestWidget.foreground': '#14171f',
        'editorSuggestWidget.border': '#dbe2f0',
        'editorHoverWidget.foreground': '#14171f',
        'editorHoverWidget.border': '#dbe2f0',
        'editor.hoverHighlightBackground': '#343a4d',
    },
    rules: [
        ...lightRules,
        // Sourcegraph base language tokens
        { token: 'identifier', foreground: '#14171f' },
        { token: 'field', foreground: '#0b70db' },
        { token: 'keyword', foreground: '#a305e1' },
        { token: 'openingParen', foreground: '#a305e1' },
        { token: 'closingParen', foreground: '#a305e1' },
        // Sourcegraph decorated language tokens
        { token: 'metaRepoRevisionSeparator', foreground: '#0b70db' },
        { token: 'metaContextPrefix', foreground: '#a305e1' },
        { token: 'metaPredicateNameAccess', foreground: '#a305e1' },
        { token: 'metaPredicateDot', foreground: '#14171f' },
        // Regexp pattern highlighting
        { token: 'metaRegexpCharacterSet', foreground: '#a305e1' },
        { token: 'metaRegexpCharacterClass', foreground: '#a305e1' },
        { token: 'metaRegexpCharacterClassMember', foreground: '#14171f' },
        { token: 'metaRegexpCharacterClassRange', foreground: '#14171f' },
        { token: 'metaRegexpCharacterClassRangeHyphen', foreground: '#a305e1' },
        // Structural pattern highlighting
        { token: 'metaStructuralVariable', foreground: '#14171f' },
        // Revision highlighting
        { token: 'metaRevisionCommitHash', foreground: '#14171f' },
        { token: 'metaRevisionLabel', foreground: '#14171f' },
        { token: 'metaRevisionReferencePath', foreground: '#14171f' },
    ],
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

    /** Whether to use the redesign theme */
    isRedesignEnabled?: boolean
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
            theme: this.getTheme(this.props.isLightTheme, this.props.isRedesignEnabled),
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
                    map(({ isLightTheme, isRedesignEnabled }) => this.getTheme(isLightTheme, isRedesignEnabled)),
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

    private getTheme = (isLightTheme: boolean, isRedesignEnabled?: boolean): string =>
        isRedesignEnabled
            ? isLightTheme
                ? SOURCEGRAPH_LIGHT_REDESIGN
                : SOURCEGRAPH_DARK_REDESIGN
            : isLightTheme
            ? SOURCEGRAPH_LIGHT
            : SOURCEGRAPH_DARK
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
