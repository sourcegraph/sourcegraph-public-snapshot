import React from 'react'
import * as H from 'history'
import * as Monaco from 'monaco-editor'
import { MonacoEditor } from '../../components/MonacoEditor'
import { QueryState } from '../helpers'
import { getProviders } from '../../../../shared/src/search/parser/providers'
import { Subscription, Observable, Subject, Unsubscribable } from 'rxjs'
import { fetchSuggestions } from '../backend'
import { toArray, map, distinctUntilChanged, publishReplay, refCount } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { ThemeProps } from '../../../../shared/src/theme'
import { CaseSensitivityProps, PatternTypeProps } from '..'
import { Toggles, TogglesProps } from './toggles/Toggles'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'

export interface MonacoQueryInputProps
    extends Omit<TogglesProps, 'navbarSearchQuery'>,
        ThemeProps,
        CaseSensitivityProps,
        PatternTypeProps {
    location: H.Location
    history: H.History
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    autoFocus?: boolean
}

const SOURCEGRAPH_SEARCH: 'sourcegraphSearch' = 'sourcegraphSearch'

type Theme = 'sourcegraph-dark' | 'sourcegraph-light'

/**
 * Maps a Monaco IDisposable to an rxjs Unsubscribable.
 */
const toUnsubscribable = (disposable: Monaco.IDisposable): Unsubscribable => ({
    unsubscribe: () => disposable.dispose(),
})

/**
 * Adds code intelligence for the Sourcegraph search syntax to Monaco.
 *
 * @returns Subscription
 */
function addSouregraphSearchCodeIntelligence(
    monaco: typeof Monaco,
    searchQueries: Observable<string>,
    patternTypes: Observable<SearchPatternType>,
    themeChanges: Observable<Theme>
): Subscription {
    const subscriptions = new Subscription()

    // Register language ID
    monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

    // Register themes and handle theme change
    monaco.editor.defineTheme('sourcegraph-dark', {
        base: 'vs-dark',
        inherit: false,
        colors: {
            background: '#0E121B',
            'textLink.activeBackground': '#2a3a51',
            'editor.background': '#0E121B',
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
        },
        rules: [
            { token: 'identifier', foreground: '#f2f4f8' },
            { token: 'keyword', foreground: '#569cd6' },
        ],
    })
    monaco.editor.defineTheme('sourcegraph-light', {
        base: 'vs',
        inherit: false,
        colors: {
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
        },
        rules: [
            { token: 'identifier', foreground: '#2b3750' },
            { token: 'keyword', foreground: '#268bd2' },
        ],
    })
    subscriptions.add(
        themeChanges.subscribe(theme => {
            monaco.editor.setTheme(theme)
        })
    )

    // Register providers
    const providers = getProviders(searchQueries, patternTypes, (query: string) =>
        fetchSuggestions(query).pipe(toArray())
    )
    subscriptions.add(toUnsubscribable(monaco.languages.setTokensProvider(SOURCEGRAPH_SEARCH, providers.tokens)))
    subscriptions.add(toUnsubscribable(monaco.languages.registerHoverProvider(SOURCEGRAPH_SEARCH, providers.hover)))
    subscriptions.add(
        toUnsubscribable(monaco.languages.registerCompletionItemProvider(SOURCEGRAPH_SEARCH, providers.completion))
    )
    subscriptions.add(
        providers.diagnostics.subscribe(markers => {
            monaco.editor.setModelMarkers(monaco.editor.getModels()[0], 'diagnostics', markers)
        })
    )

    return subscriptions
}

/**
 * A search query input backed by the Monaco editor, allowing it to provide
 * syntax highlighting, hovers, completions and diagnostics for search queries.
 *
 * This component should not be imported directly: use {@link LazyMonacoQueryInput} instead
 * to avoid bundling the Monaco editor on every page.
 */
export class MonacoQueryInput extends React.PureComponent<MonacoQueryInputProps> {
    private componentUpdates = new Subject<MonacoQueryInputProps>()
    private searchQueries = this.componentUpdates.pipe(
        map(({ queryState }) => queryState.query),
        distinctUntilChanged(),
        publishReplay(1),
        refCount()
    )
    private themeChanges = this.componentUpdates.pipe(
        map(({ isLightTheme }): Theme => (isLightTheme ? 'sourcegraph-light' : 'sourcegraph-dark')),
        distinctUntilChanged(),
        publishReplay(1),
        refCount()
    )
    private patternTypes = this.componentUpdates.pipe(
        map(({ patternType }) => patternType),
        distinctUntilChanged(),
        publishReplay(1),
        refCount()
    )
    private containerRef: HTMLElement | null = null
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        const options: Monaco.editor.IEditorOptions = {
            readOnly: false,
            lineNumbers: 'off',
            lineHeight: 16,
            // Match the query input's height for suggestion items line height.
            suggestLineHeight: 34,
            minimap: {
                enabled: false,
            },
            scrollbar: {
                vertical: 'hidden',
                horizontal: 'hidden',
            },
            glyphMargin: false,
            lineDecorationsWidth: 0,
            lineNumbersMinChars: 0,
            overviewRulerBorder: false,
            folding: false,
            rulers: [],
            overviewRulerLanes: 0,
            wordBasedSuggestions: false,
            quickSuggestions: false,
            fixedOverflowWidgets: true,
            contextmenu: false,
            // Display the cursor as a 1px line.
            cursorStyle: 'line',
            cursorWidth: 1,
        }
        return (
            <div ref={this.setContainerRef} className="monaco-query-input-container flex-1">
                <div className="flex-1">
                    <MonacoEditor
                        id="monaco-query-input"
                        language={SOURCEGRAPH_SEARCH}
                        value={this.props.queryState.query}
                        height={16}
                        theme="sourcegraph-dark"
                        editorWillMount={this.editorWillMount}
                        onEditorCreated={this.onEditorCreated}
                        options={options}
                        border={false}
                    ></MonacoEditor>
                </div>
                <Toggles {...this.props} navbarSearchQuery={this.props.queryState.query} />
            </div>
        )
    }

    private setContainerRef = (ref: HTMLElement | null): void => {
        this.containerRef = ref
    }

    private onChange = (query: string): void => {
        // Cursor position is irrelevant for the Monaco query input.
        this.props.onChange({ query, cursorPosition: 0 })
    }

    private onSubmit = (): void => {
        this.props.onSubmit()
    }

    private editorWillMount = (monaco: typeof Monaco): void => {
        // Register themes and code intelligence providers.
        this.subscriptions.add(
            addSouregraphSearchCodeIntelligence(monaco, this.searchQueries, this.patternTypes, this.themeChanges)
        )
    }

    private onEditorCreated = (editor: Monaco.editor.IStandaloneCodeEditor): void => {
        if (this.props.autoFocus) {
            // Focus the editor with cursor at end.
            editor.focus()
            editor.setPosition({
                // +2 as Monaco is 1-indexed, and the cursor should be placed after the query.
                column: editor.getValue().length + 2,
                lineNumber: 1,
            })
        }
        // Prevent newline insertion in model, and surface query changes with stripped newlines.
        this.subscriptions.add(
            toUnsubscribable(
                editor.onDidChangeModelContent(() => {
                    this.onChange(editor.getValue().replace(/[\n\r↵]/g, ''))
                })
            )
        )

        // Submit on enter when not showing suggestions.
        this.subscriptions.add(
            toUnsubscribable(
                editor.addAction({
                    id: 'submitOnEnter',
                    label: 'submitOnEnter',
                    keybindings: [Monaco.KeyCode.Enter],
                    precondition: '!suggestWidgetVisible',
                    run: () => {
                        this.onSubmit()
                    },
                })
            )
        )
        // Prevent inserting newlines.
        this.subscriptions.add(
            toUnsubscribable(
                editor.onKeyDown(e => {
                    if (e.keyCode === Monaco.KeyCode.Enter) {
                        e.preventDefault()
                    }
                })
            )
        )
        // Trigger a layout of the Monaco editor when its container gets resized.
        // The Monaco editor doesn't auto-resize with its container:
        // https://github.com/microsoft/monaco-editor/issues/28
        if (this.containerRef) {
            const resizeObserver = new ResizeObserver(() => {
                editor.layout()
            })
            resizeObserver.observe(this.containerRef)
            this.subscriptions.add(() => resizeObserver.disconnect())
        }
    }
}
