import React from 'react'
import * as H from 'history'
import * as Monaco from 'monaco-editor'
import { isEqual, isPlainObject } from 'lodash'
import { MonacoEditor } from '../../components/MonacoEditor'
import { QueryState } from '../helpers'
import { getProviders } from '../../../../shared/src/search/query/providers'
import { Subscription, Observable, Subject, Unsubscribable, ReplaySubject } from 'rxjs'
import { fetchSuggestions } from '../backend'
import { map, distinctUntilChanged, filter, switchMap, withLatestFrom } from 'rxjs/operators'
import { Omit } from 'utility-types'
import { ThemeProps } from '../../../../shared/src/theme'
import { CaseSensitivityProps, PatternTypeProps, CopyQueryButtonProps } from '..'
import { Toggles, TogglesProps } from './toggles/Toggles'
import { hasProperty, isDefined } from '../../../../shared/src/util/types'
import { KeyboardShortcut } from '../../../../shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '../../keyboardShortcuts/keyboardShortcuts'
import { observeResize } from '../../util/dom'
import { SearchPatternType } from '../../graphql-operations'

export interface MonacoQueryInputProps
    extends Omit<TogglesProps, 'navbarSearchQuery'>,
        ThemeProps,
        CaseSensitivityProps,
        PatternTypeProps,
        CopyQueryButtonProps {
    location: H.Location
    history: H.History
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit: () => void
    onFocus?: () => void
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut

    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether to additionally highlight or provide hovers for tokens, e.g., regexp character sets.
    enableSmartQuery: boolean

    // Whether comments are parsed and highlighted
    interpretComments?: boolean
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
export function addSourcegraphSearchCodeIntelligence(
    monaco: typeof Monaco,
    searchQueries: Observable<string>,
    options: {
        patternType: SearchPatternType
        globbing: boolean
        interpretComments?: boolean
        enableSmartQuery: boolean
    }
): Subscription {
    const subscriptions = new Subscription()

    // Register language ID
    monaco.languages.register({ id: SOURCEGRAPH_SEARCH })

    // Register providers
    const providers = getProviders(searchQueries, fetchSuggestions, options)
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
 * HACK: this interface and the below type guard are used to add a custom command
 * to the editor. There is no public API to add a command with a specified ID and handler,
 * hence we need to use the private _commandService API.
 *
 * See upstream issue:
 * - https://github.com/Microsoft/monaco-editor/issues/900#issue-327455729
 * */
interface MonacoEditorWithCommandService extends Monaco.editor.IStandaloneCodeEditor {
    _commandService: {
        addCommand: (command: { id: string; handler: () => void }) => void
    }
}

const hasCommandService = (editor: Monaco.editor.IStandaloneCodeEditor): editor is MonacoEditorWithCommandService =>
    hasProperty('_commandService')(editor) &&
    typeof (editor._commandService as MonacoEditorWithCommandService['_commandService']).addCommand === 'function'

/**
 * A search query input backed by the Monaco editor, allowing it to provide
 * syntax highlighting, hovers, completions and diagnostics for search queries.
 *
 * This component should not be imported directly: use {@link LazyMonacoQueryInput} instead
 * to avoid bundling the Monaco editor on every page.
 */
export class MonacoQueryInput extends React.PureComponent<MonacoQueryInputProps> {
    private componentUpdates = new ReplaySubject<MonacoQueryInputProps>(1)
    private searchQueries = this.componentUpdates.pipe(
        map(({ queryState }) => queryState.query),
        distinctUntilChanged()
    )
    private containerRefs = new Subject<HTMLElement | null>()
    private editorRefs = new Subject<Monaco.editor.IStandaloneCodeEditor | null>()
    private subscriptions = new Subscription()

    constructor(props: MonacoQueryInputProps) {
        super(props)
        // Trigger a layout of the Monaco editor when its container gets resized.
        // The Monaco editor doesn't auto-resize with its container:
        // https://github.com/microsoft/monaco-editor/issues/28
        this.subscriptions.add(
            this.containerRefs
                .pipe(
                    switchMap(container => (container ? observeResize(container) : [])),
                    withLatestFrom(this.editorRefs),
                    map(([, editor]) => editor),
                    filter(isDefined)
                )
                .subscribe(editor => {
                    editor.layout()
                })
        )
    }

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
            links: false,
            // Display the cursor as a 1px line.
            cursorStyle: 'line',
            cursorWidth: 1,
        }
        return (
            <>
                <div ref={this.containerRefs.next.bind(this.containerRefs)} className="monaco-query-input-container">
                    <div className="flex-grow-1 flex-shrink-past-contents" onFocus={this.props.onFocus}>
                        <MonacoEditor
                            id="monaco-query-input"
                            language={SOURCEGRAPH_SEARCH}
                            value={this.props.queryState.query}
                            height={17}
                            isLightTheme={this.props.isLightTheme}
                            editorWillMount={this.editorWillMount}
                            onEditorCreated={this.onEditorCreated}
                            options={options}
                            border={false}
                            keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                            className="test-query-input"
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

    private onChange = (query: string): void => {
        this.props.onChange({ query, fromUserInput: true })
    }

    private onSubmit = (): void => {
        this.props.onSubmit()
    }

    private editorWillMount = (monaco: typeof Monaco): void => {
        // Register themes and code intelligence providers.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ patternType, globbing, enableSmartQuery, interpretComments }) => ({
                        patternType,
                        globbing,
                        enableSmartQuery,
                        interpretComments,
                    })),
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(
                        options =>
                            new Observable(() =>
                                addSourcegraphSearchCodeIntelligence(monaco, this.searchQueries, options)
                            )
                    )
                )
                .subscribe()
        )
    }

    private onEditorCreated = (editor: Monaco.editor.IStandaloneCodeEditor): void => {
        this.props.onSuggestionsInitialized?.({
            trigger: () => editor.trigger('triggerSuggestions', 'editor.action.triggerSuggest', {}),
        })

        if (this.props.onCompletionItemSelected) {
            if (!hasCommandService(editor)) {
                throw new Error('Could not call onCompletionItemSelected: editor has no commandService')
            }
            this.subscriptions.add(
                editor._commandService.addCommand({
                    id: 'completionItemSelected',
                    handler: this.props.onCompletionItemSelected,
                })
            )
        }

        this.editorRefs.next(editor)
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

        // Focus the editor if the autoFocus prop is truthy
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ autoFocus }) => autoFocus),
                    filter(autoFocus => !!autoFocus),
                    distinctUntilChanged()
                )
                .subscribe(() => {
                    editor.focus()
                })
        )

        // If an edit wasn't triggered by the user,
        // place the cursor at the end of the query.
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ queryState }) => queryState),
                    filter(({ fromUserInput }) => !fromUserInput),
                    distinctUntilChanged((a, b) => a.query === b.query)
                )
                .subscribe(() => {
                    const position = {
                        // +2 as Monaco is 1-indexed.
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
            // Keep ctrl+space to show all available completions. Keep ctrl+k to delete text on right of cursor.
            if (action === 'editor.action.triggerSuggest' || action === 'deleteAllRight') {
                continue
            }
            // Prefixing action ids with `-` to unbind the default actions.
            editor._standaloneKeybindingService.addDynamicKeybinding(`-${action}`)
        }
        // Free CMD+L keybinding, which is part of Monaco's CoreNavigationCommands, and
        // not exposed on editor._actions.
        editor._standaloneKeybindingService.addDynamicKeybinding('-expandLineSelection')
    }
}
