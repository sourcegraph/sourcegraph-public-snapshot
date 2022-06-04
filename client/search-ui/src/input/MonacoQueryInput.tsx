import React, { useCallback, useEffect, useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { isPlainObject, noop } from 'lodash'
import * as Monaco from 'monaco-editor'

import { observeResize, hasProperty } from '@sourcegraph/common'
import {
    QueryChangeSource,
    QueryState,
    CaseSensitivityProps,
    SearchPatternTypeProps,
    SearchContextProps,
} from '@sourcegraph/search'
import { MonacoEditor } from '@sourcegraph/shared/src/components/MonacoEditor'
import { KeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts'
import { KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR } from '@sourcegraph/shared/src/keyboardShortcuts/keyboardShortcuts'
import { toMonacoRange } from '@sourcegraph/shared/src/search/query/monaco'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { IEditor } from './LazyMonacoQueryInput'
import { useQueryDiagnostics, useQueryIntelligence } from './useQueryIntelligence'

import styles from './MonacoQueryInput.module.scss'

export const DEFAULT_MONACO_OPTIONS: Monaco.editor.IStandaloneEditorConstructionOptions = {
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
    hover: { delay: 150 },
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
    // Match our monospace/code style from code.scss
    fontFamily: 'sfmono-regular, consolas, menlo, dejavu sans mono, monospace',
    // Display the cursor as a 1px line.
    cursorStyle: 'line',
    cursorWidth: 1,
}

export interface MonacoQueryInputProps
    extends ThemeProps,
        Pick<CaseSensitivityProps, 'caseSensitive'>,
        SearchPatternTypeProps,
        Pick<SearchContextProps, 'selectedSearchContextSpec'> {
    isSourcegraphDotCom: boolean // Needed for query suggestions to give different options on dotcom; see SOURCEGRAPH_DOT_COM_REPO_COMPLETION
    queryState: QueryState
    onChange: (newState: QueryState) => void
    onSubmit?: () => void
    onFocus?: () => void
    onBlur?: () => void
    onCompletionItemSelected?: () => void
    onSuggestionsInitialized?: (actions: { trigger: () => void }) => void
    onEditorCreated?: (editor: IEditor) => void
    fetchStreamSuggestions?: typeof defaultFetchStreamSuggestions // Alternate implementation is used in the VS Code extension.
    autoFocus?: boolean
    keyboardShortcutForFocus?: KeyboardShortcut
    onHandleFuzzyFinder?: React.Dispatch<React.SetStateAction<boolean>>
    // Whether globbing is enabled for filters.
    globbing: boolean

    // Whether comments are parsed and highlighted
    interpretComments?: boolean

    className?: string

    height?: string | number
    preventNewLine?: boolean
    editorOptions?: Monaco.editor.IStandaloneEditorConstructionOptions

    /**
     * NOTE: This is currently only used for Insights code through
     * the MonacoField component: client/web/src/enterprise/insights/components/form/monaco-field/MonacoField.tsx
     *
     * Issue to improve this: https://github.com/sourcegraph/sourcegraph/issues/29438
     */
    placeholder?: string

    ariaLabel?: string

    editorClassName?: string
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
        addDynamicKeybinding(
            commandId: string,
            _keybinding: number | undefined,
            handler: Monaco.editor.ICommandHandler
        ): void
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

export const toMonacoSelection = (range: Monaco.IRange): Monaco.ISelection => ({
    selectionStartLineNumber: range.startLineNumber,
    positionLineNumber: range.endLineNumber,
    selectionStartColumn: range.startColumn,
    positionColumn: range.endColumn,
})

/**
 * A search query input backed by the Monaco editor, allowing it to provide
 * syntax highlighting, hovers, completions and diagnostics for search queries.
 *
 * This component should not be imported directly: use {@link LazyMonacoQueryInput} instead
 * to avoid bundling the Monaco editor on every page.
 */
export const MonacoQueryInput: React.FunctionComponent<React.PropsWithChildren<MonacoQueryInputProps>> = ({
    queryState,
    onFocus,
    onBlur,
    onChange,
    onSubmit = noop,
    onSuggestionsInitialized,
    onCompletionItemSelected,
    fetchStreamSuggestions = defaultFetchStreamSuggestions,
    autoFocus,
    selectedSearchContextSpec,
    patternType,
    globbing,
    interpretComments,
    isSourcegraphDotCom,
    isLightTheme,
    className,
    height = 17,
    preventNewLine = true,
    editorOptions,
    onHandleFuzzyFinder,
    editorClassName,
    onEditorCreated: onEditorCreatedCallback,
    placeholder,
    ariaLabel = 'Search query',
}) => {
    const [editor, setEditor] = useState<Monaco.editor.IStandaloneCodeEditor>()

    const onEditorCreated = useCallback(
        (editor: Monaco.editor.IStandaloneCodeEditor) => {
            // `role` set to fix accessibility issues
            // https://github.com/sourcegraph/sourcegraph/issues/34733
            editor.getDomNode()?.setAttribute('role', 'textbox')
            // `aria-label` to fix accessibility audit
            editor.getDomNode()?.setAttribute('aria-label', ariaLabel)

            setEditor(editor)
            onEditorCreatedCallback?.(editor)
        },
        [setEditor, onEditorCreatedCallback, ariaLabel]
    )

    // Trigger a layout of the Monaco editor when its container gets resized.
    // The Monaco editor doesn't auto-resize with its container:
    // https://github.com/microsoft/monaco-editor/issues/28
    const [container, setContainer] = useState<HTMLElement | null>()

    useLayoutEffect(() => {
        if (!editor || !container) {
            return
        }
        const subscription = observeResize(container).subscribe(() => {
            editor.layout()
        })
        return () => subscription.unsubscribe()
    }, [editor, container])

    const fetchSuggestionsWithContext = useCallback(
        (query: string) => fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpec)),
        [selectedSearchContextSpec, fetchStreamSuggestions]
    )

    const sourcegraphSearchLanguageId = useQueryIntelligence(fetchSuggestionsWithContext, {
        patternType,
        globbing,
        interpretComments,
        isSourcegraphDotCom,
    })

    useQueryDiagnostics(editor, { patternType, interpretComments })

    // Register suggestions handle
    useEffect(() => {
        if (editor) {
            onSuggestionsInitialized?.({
                trigger: () => editor.trigger('triggerSuggestions', 'editor.action.triggerSuggest', {}),
            })
        }
    }, [editor, onSuggestionsInitialized])

    // Register onCompletionSelected handler
    useEffect(() => {
        const disposable = Monaco.editor.registerCommand('completionItemSelected', onCompletionItemSelected ?? noop)
        return () => disposable.dispose()
    }, [onCompletionItemSelected])

    // Disable default Monaco keybindings
    useEffect(() => {
        if (!editor) {
            return
        }
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
            editor._standaloneKeybindingService.addDynamicKeybinding(`-${action}`, undefined, () => {})
        }
        // Free CMD+L keybinding, which is part of Monaco's CoreNavigationCommands, and
        // not exposed on editor._actions.
        editor._standaloneKeybindingService.addDynamicKeybinding('-expandLineSelection', undefined, () => {})
    }, [editor])

    // Accessibility: allow tab usage to move focus to
    // next previous focusable element (and not to insert the tab character).
    // - Cannot be set through IEditorOptions
    // - Cannot be called synchronously (otherwise risks being overridden by Monaco defaults)
    useEffect(() => {
        if (!editor) {
            return
        }
        const disposable = editor.onDidFocusEditorText(() => {
            // Use a small non-zero timeout to avoid being overridden by Monaco defaults when editor is shown/hidden quickly
            setTimeout(() => editor.createContextKey('editorTabMovesFocus', true), 50)
        })
        return () => disposable.dispose()
    }, [editor])

    // Focus the editor if the autoFocus prop is truthy
    useEffect(() => {
        if (!editor || !autoFocus) {
            return
        }
        editor.focus()
    }, [editor, autoFocus])

    useEffect(() => {
        if (!editor || !onBlur) {
            return
        }

        const disposable = editor.onDidBlurEditorText(onBlur)

        return () => disposable.dispose()
    }, [editor, onBlur])

    // Always focus the editor on selectedSearchContextSpec change
    useEffect(() => {
        if (selectedSearchContextSpec) {
            editor?.focus()
        }
    }, [editor, selectedSearchContextSpec])

    useEffect(() => {
        if (!editor) {
            return
        }

        switch (queryState.changeSource) {
            case QueryChangeSource.userInput:
                // Don't react to user input
                break
            case QueryChangeSource.searchTypes:
            case QueryChangeSource.searchReference: {
                const textModel = editor.getModel()
                if (textModel) {
                    const selectionRange = toMonacoSelection(toMonacoRange(queryState.selectionRange, textModel))
                    editor.setSelection(selectionRange)
                    if (queryState.showSuggestions) {
                        editor.trigger('triggerSuggestions', 'editor.action.triggerSuggest', {})
                    }
                    // For some reason this has to come *after* triggering the
                    // suggestion, otherwise the suggestion box will be shown
                    // and the filter is not scrolled into view.
                    editor.revealRange(toMonacoRange(queryState.revealRange, textModel))
                }
                editor.focus()
                break
            }
            default: {
                // Place the cursor at the end of the query.
                const position = {
                    // +2 as Monaco is 1-indexed.
                    column: editor.getValue().length + 2,
                    lineNumber: 1,
                }
                editor.setPosition(position)
                editor.revealPosition(position)
            }
        }
    }, [editor, queryState])

    // Prevent newline insertion in model, and surface query changes with stripped newlines.
    useEffect(() => {
        if (!editor) {
            return
        }
        const replacePattern = /[\n\r↵]/g
        const disposable = editor.onDidChangeModelContent(() => {
            const value = editor.getValue()
            onChange({
                query: preventNewLine ? value.replace(replacePattern, '') : value,
                changeSource: QueryChangeSource.userInput,
            })
        })
        return () => disposable.dispose()
    }, [editor, onChange, preventNewLine])

    // Submit on enter, hiding the suggestions widget if it's visible.
    useEffect(() => {
        if (!editor) {
            return
        }

        // Unconditionally trigger the search when pressing `Enter`,
        // including when there are visible completion suggestions.
        const disposables = [
            editor.addAction({
                id: 'submitOnEnter',
                label: 'submitOnEnter',
                keybindings: [Monaco.KeyCode.Enter],
                run: () => {
                    onSubmit()
                    editor.trigger('submitOnEnter', 'hideSuggestWidget', [])
                },
            }),
        ]

        if (onHandleFuzzyFinder) {
            disposables.push(
                editor.addAction({
                    id: 'triggerFuzzyFinder',
                    label: 'triggerFuzzyFinder',
                    keybindings: [Monaco.KeyMod.CtrlCmd | Monaco.KeyCode.KEY_K],
                    run: () => onHandleFuzzyFinder(true),
                })
            )
        }

        return () => {
            for (const disposable of disposables) {
                disposable.dispose()
            }
        }
    }, [editor, onSubmit, onHandleFuzzyFinder])

    return (
        <div
            ref={setContainer}
            className={classNames('flex-grow-1 flex-shrink-past-contents', className)}
            onFocus={onFocus}
        >
            <MonacoEditor
                id="monaco-query-input"
                language={sourcegraphSearchLanguageId}
                value={queryState.query}
                height={height}
                isLightTheme={isLightTheme}
                editorWillMount={noop}
                onEditorCreated={onEditorCreated}
                options={editorOptions ?? DEFAULT_MONACO_OPTIONS}
                border={false}
                placeholder={placeholder}
                keyboardShortcutForFocus={KEYBOARD_SHORTCUT_FOCUS_SEARCHBAR}
                className={classNames('test-query-input', styles.monacoQueryInput, editorClassName)}
            />
        </div>
    )
}
