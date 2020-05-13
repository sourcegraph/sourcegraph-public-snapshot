import React from 'react'
import * as H from 'history'
import * as Monaco from 'monaco-editor'
import { isPlainObject } from 'lodash'
import { MonacoEditor } from '../../components/MonacoEditor'
import { QueryState } from '../helpers'
import { getProviders } from '../../../../shared/src/search/parser/providers'
import { Subscription, Observable, Subject, Unsubscribable } from 'rxjs'
import { fetchSuggestions } from '../backend'
import { map, distinctUntilChanged, publishReplay, refCount, filter } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { ThemeProps } from '../../../../shared/src/theme'
import { CaseSensitivityProps, PatternTypeProps } from '..'
import { Toggles, TogglesProps } from './toggles/Toggles'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { hasProperty } from '../../../../shared/src/util/types'
import { KeyboardShortcut } from '../../../../shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'

export interface MonacoQueryInputProps
    extends Omit<TogglesProps, 'navbarSearchQuery' | 'filtersInQuery'>,
        ThemeProps,
        CaseSensitivityProps,
        PatternTypeProps {
    location: H.Location
    history: H.History
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut
}

const SOURCEGRAPH_SEARCH = 'sourcegraphSearch' as const

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
    patternTypes: Observable<SearchPatternType>
): Subscription {
    const subscriptions = new Subscription()

    // Register language ID
    monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

    // Register providers
    const providers = getProviders(searchQueries, patternTypes, fetchSuggestions)
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
 * HACK: this interface and the below type guard are used to free default Monaco
 * keybindings (such as cmd + F, cmd + L) by unregistering them from the private
 * `_standaloneKeybindingService`.
 *
 * This is necessary as simply registering a noop command with editor.addCommand(keybinding, noop)
 * prevents the default Monaco behaviour, but doesn't free the keybinding, and thus still blocks the
 * default browser action (eg. select location with cmd + L).
 *
 * See upstream issues:
 * - https://github.com/microsoft/monaco-editor/issues/287
 * - https://github.com/microsoft/monaco-editor/issues/102 (main tracking issue)
 */
interface MonacoEditorWithKeybindingsService extends Monaco.editor.IStandaloneCodeEditor {
    _actions: {
        [id: string]: {
            id: string
            alias: string
            label: string
        }
    }
    _standaloneKeybindingService: {
        addDynamicKeybinding(keybinding: string): void
    }
}

const hasKeybindingService = (
    editor: Monaco.editor.IStandaloneCodeEditor
): editor is MonacoEditorWithKeybindingsService =>
    hasProperty('_actions')(editor) &&
    isPlainObject(editor._actions) &&
    hasProperty('_standaloneKeybindingService')(editor) &&
    typeof (editor._standaloneKeybindingService as MonacoEditorWithKeybindingsService['_standaloneKeybindingService'])
        .addDynamicKeybinding === 'function'

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
            <>
                <div ref={this.setContainerRef} className="monaco-query-input-container flex-1">
                    <div className="flex-1">
                        <MonacoEditor
                            id="monaco-query-input"
                            language={SOURCEGRAPH_SEARCH}
                            value={this.props.queryState.query}
                            height={16}
                            isLightTheme={this.props.isLightTheme}
                            editorWillMount={this.editorWillMount}
                            onEditorCreated={this.onEditorCreated}
                            options={options}
                            border={false}
                            keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                        />
                    </div>
                    <Toggles
                        {...this.props}
                        navbarSearchQuery={this.props.queryState.query}
                        className="monaco-query-input-container__toggle-container"
                    />
                </div>
            </>
        )
    }

    private setContainerRef = (ref: HTMLElement | null): void => {
        this.containerRef = ref
    }

    private onChange = (query: string): void => {
        // Cursor position is irrelevant for the Monaco query input.
        this.props.onChange({ query, cursorPosition: 0, fromUserInput: true })
    }

    private onSubmit = (): void => {
        this.props.onSubmit()
    }

    private editorWillMount = (monaco: typeof Monaco): void => {
        // Register themes and code intelligence providers.
        this.subscriptions.add(addSouregraphSearchCodeIntelligence(monaco, this.searchQueries, this.patternTypes))
    }

    private onEditorCreated = (editor: Monaco.editor.IStandaloneCodeEditor): void => {
        // Accessibility: allow tab usage to move focus to
        // next previous focusable element (and not to insert the tab character).
        // - Cannot be set through IEditorOptions
        // - Cannot be called synchronously (otherwise risks being overridden by Monaco defaults)
        this.subscriptions.add(
            toUnsubscribable(
                editor.onDidFocusEditorText(() => {
                    editor.createContextKey('editorTabMovesFocus', true)
                })
            )
        )

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    filter(({ autoFocus }) => !!autoFocus),
                    map(({ queryState }) => queryState),
                    filter(({ fromUserInput }) => !fromUserInput),
                    distinctUntilChanged((a, b) => a.query === b.query)
                )
                .subscribe(() => {
                    // Focus the editor with cursor at end, and reveal that position.
                    editor.focus()
                    const position = {
                        // +2 as Monaco is 1-indexed, and the cursor should be placed after the query.
                        column: editor.getValue().length + 2,
                        lineNumber: 1,
                    }
                    editor.setPosition(position)
                    editor.revealPosition(position)
                })
        )

        // Prevent newline insertion in model, and surface query changes with stripped newlines.
        this.subscriptions.add(
            toUnsubscribable(
                editor.onDidChangeModelContent(() => {
                    this.onChange(editor.getValue().replace(/[\n\r↵]/g, ''))
                })
            )
        )

        // Submit on enter, hiding the suggestions widget if it's visible.
        this.subscriptions.add(
            toUnsubscribable(
                editor.addAction({
                    id: 'submitOnEnter',
                    label: 'submitOnEnter',
                    keybindings: [Monaco.KeyCode.Enter],
                    run: () => {
                        this.onSubmit()
                        editor.trigger('submitOnEnter', 'hideSuggestWidget', [])
                    },
                })
            )
        )

        // Disable default Monaco keybindings
        if (!hasKeybindingService(editor)) {
            // Throw an error if hasKeybindingService() returns false,
            // to surface issues with this workaround when upgrading Monaco.
            throw new Error('Cannot unbind default Monaco keybindings')
        }
        for (const action of Object.keys(editor._actions)) {
            // Prefixing action ids with `-` to unbind the default actions.
            editor._standaloneKeybindingService.addDynamicKeybinding(`-${action}`)
        }
        // Free CMD+L keybinding, which is part of Monaco's CoreNavigationCommands, and
        // not exposed on editor._actions.
        editor._standaloneKeybindingService.addDynamicKeybinding('-expandLineSelection')

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
