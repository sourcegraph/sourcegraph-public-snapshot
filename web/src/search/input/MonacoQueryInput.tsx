import React from 'react'
import * as Monaco from 'monaco-editor'
import { MonacoEditor } from '../../components/MonacoEditor'
import { QueryState } from '../helpers'
import { getProviders } from '../parser/providers'
import { Subscription, Observable, Subject, Unsubscribable } from 'rxjs'
import { fetchSuggestions } from '../backend'
import { toArray, map, distinctUntilChanged, publishReplay, refCount } from 'rxjs/operators'
import { RegexpToggle, RegexpToggleProps } from './RegexpToggle'
import { Omit } from 'utility-types'
import { ThemeProps } from '../../../../shared/src/theme'

export interface MonacoQueryInputProps extends Omit<RegexpToggleProps, 'navbarSearchQuery'>, ThemeProps {
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
    themeChanges: Observable<Theme>
): Subscription {
    const subscriptions = new Subscription()

    // Register language ID
    monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

    // Register themes and handle theme change
    monaco.editor.defineTheme('sourcegraph-dark', {
        base: 'vs-dark',
        inherit: true,
        colors: {
            'editor.background': '#0E121B',
            'editor.foreground': '#ffffff',
            'editorCursor.foreground': '#ffffff',
            'editor.selectionBackground': '#1C7CD650',
            'editor.selectionHighlightBackground': '#1C7CD625',
            'editor.inactiveSelectionBackground': '#1C7CD625',
            'editorSuggestWidget.background': '#1c2736',
            'editorSuggestWidget.foreground': '#F2F4F8',
            'editorSuggestWidget.border': '#2b3750',
            'editorHoverWidget.background': '#1c2736',
            'editorHoverWidget.foreground': '#F2F4F8',
            'editorHoverWidget.border': '#2b3750',
        },
        rules: [],
    })
    monaco.editor.defineTheme('sourcegraph-light', {
        base: 'vs',
        inherit: true,
        colors: {
            'editor.background': '#ffffff',
            'editor.foreground': '#2b3750',
            'editorCursor.foreground': '#2b3750',
            'editor.selectionBackground': '#1C7CD650',
            'editor.selectionHighlightBackground': '#1C7CD625',
            'editor.inactiveSelectionBackground': '#1C7CD625',
            'editorSuggestWidget.background': '#ffffff',
            'editorSuggestWidget.foreground': '#2b3750',
            'editorSuggestWidget.border': '#cad2e2',
            'editorHoverWidget.background': '#ffffff',
            'editorHoverWidget.foreground': '#2b3750',
            'editorHoverWidget.border': '#cad2e2',
        },
        rules: [],
    })
    subscriptions.add(
        themeChanges.subscribe(theme => {
            monaco.editor.setTheme(theme)
        })
    )

    // Register providers
    const providers = getProviders(searchQueries, (query: string) => fetchSuggestions(query).pipe(toArray()))
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
            lineHeight: 32,
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
        }
        return (
            <div ref={this.setContainerRef} className="monaco-query-input-container flex-1">
                <MonacoEditor
                    id="monaco-query-input"
                    language={SOURCEGRAPH_SEARCH}
                    value={this.props.queryState.query}
                    height={34}
                    theme="sourcegraph-dark"
                    editorWillMount={this.editorWillMount}
                    onEditorCreated={this.onEditorCreated}
                    options={options}
                    border={false}
                ></MonacoEditor>
                <RegexpToggle {...this.props} navbarSearchQuery={this.props.queryState.query}></RegexpToggle>
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
        this.subscriptions.add(addSouregraphSearchCodeIntelligence(monaco, this.searchQueries, this.themeChanges))
    }

    private onEditorCreated = (editor: Monaco.editor.IStandaloneCodeEditor): void => {
        // Focus the editor by default, with cursor at end.
        editor.focus()
        editor.setPosition({
            column: editor.getValue().length + 2,
            lineNumber: 1,
        })
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
        if (this.containerRef) {
            const resizeObserver = new ResizeObserver(() => {
                editor.layout()
            })
            resizeObserver.observe(this.containerRef)
            this.subscriptions.add(() => resizeObserver.disconnect())
        }
    }
}
