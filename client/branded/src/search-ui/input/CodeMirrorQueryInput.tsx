import React, { RefObject, useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { closeCompletion, startCompletion } from '@codemirror/autocomplete'
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands'
import { EditorSelection, Extension, Prec, Compartment, EditorState } from '@codemirror/state'
import { EditorView, ViewUpdate, keymap, placeholder as placeholderExtension } from '@codemirror/view'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { useCodeMirror, createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import {
    EditorHint,
    QueryChangeSource,
    type QueryState,
    type SearchPatternTypeProps,
} from '@sourcegraph/shared/src/search'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { useIsLightTheme } from '@sourcegraph/shared/src/theme'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { createDefaultSuggestions, singleLine } from './codemirror'
import { decorateActiveFilter, filterPlaceholder } from './codemirror/active-filter'
import { queryDiagnostic } from './codemirror/diagnostics'
import { HISTORY_USER_EVENT, searchHistory as searchHistoryFacet } from './codemirror/history'
import { parseInputAsQuery, setQueryParseOptions } from './codemirror/parsedQuery'
import { querySyntaxHighlighting } from './codemirror/syntax-highlighting'
import { tokenInfo } from './codemirror/token-info'
import { QueryInputProps } from './QueryInput'

import styles from './CodeMirrorQueryInput.module.scss'

export interface CodeMirrorQueryInputFacadeProps extends QueryInputProps {
    /**
     * Used to be compatible with MonacoQueryInput's interface, but we only
     * support the `readOnly` flag.
     */
    editorOptions?: {
        readOnly?: boolean
    }

    /**
     * If set suggestions can be applied by pressing enter. In the past we
     * didn't enable this behavior because it interfered with loading
     * suggestions asynchronously, but CodeMirror allows us to disable selecting
     * a suggestion by default. This is currently an experimental feature.
     */
    applySuggestionsOnEnter?: boolean

    /**
     * When provided the query input will allow the user to "cycle" through the
     * serach history by pressing arrow up/down when the input is empty.
     */
    searchHistory?: RecentSearch[]

    /**
     * Callback to notify the parent component when the user cycles through the
     * search history.
     */
    onSelectSearchFromHistory?: () => void
}

/**
 * This component provides a drop-in replacement for MonacoQueryInput. It
 * creates the appropriate extensions and event handlers for the provided props.
 *
 * Deliberate differences compared to MonacoQueryInput:
 * - Filters are "highlighted" when the cursor is at their position
 * - Shift+Enter won't insert a new line if 'preventNewLine' is true (default)
 * - Not supplying 'onSubmit' and setting 'preventNewLine' to false will result
 * in a new line being added when Enter is pressed
 */
export const CodeMirrorMonacoFacade: React.FunctionComponent<CodeMirrorQueryInputFacadeProps> = ({
    patternType,
    selectedSearchContextSpec,
    queryState,
    onChange,
    onSubmit,
    autoFocus,
    onFocus,
    onBlur,
    isSourcegraphDotCom,
    onEditorCreated,
    interpretComments,
    className,
    preventNewLine = true,
    placeholder,
    editorOptions,
    ariaLabel = 'Search query',
    ariaLabelledby,
    ariaInvalid,
    ariaBusy,
    tabIndex = 0,
    // CodeMirror implementation specific options
    applySuggestionsOnEnter = false,
    searchHistory,
    onSelectSearchFromHistory,
    // Used by the VSCode extension (which doesn't use this component directly,
    // but added for future compatibility)
    fetchStreamSuggestions = defaultFetchStreamSuggestions,
    onCompletionItemSelected,
}) => {
    // We use both, state and a ref, for the editor instance because we need to
    // re-run some hooks when the editor changes but we also need a stable
    // reference that doesn't change across renders (and some hooks should only
    // run when a prop changes, not the editor).
    const [editor, setEditor] = useState<EditorView | undefined>()
    const editorReference = useRef<EditorView | null>(null)
    const focusSearchBarShortcut = useKeyboardShortcut('focusSearch')
    const navigate = useNavigate()

    const editorCreated = useCallback(
        (editor: EditorView) => {
            setEditor(editor)
            editorReference.current = editor
            onEditorCreated?.(editor)
        },
        [onEditorCreated]
    )

    const autocompletion = useMemo(
        () =>
            createDefaultSuggestions({
                fetchSuggestions: query =>
                    fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpec)),
                isSourcegraphDotCom,
                navigate,
                applyOnEnter: applySuggestionsOnEnter,
            }),
        [isSourcegraphDotCom, navigate, applySuggestionsOnEnter, fetchStreamSuggestions, selectedSearchContextSpec]
    )

    const extensions = useMemo(() => {
        const extensions: Extension[] = [
            EditorView.contentAttributes.of({ 'aria-label': ariaLabel }),
            callbacksField,
            autocompletion,
        ]

        if (ariaLabelledby) {
            extensions.push(EditorView.contentAttributes.of({ 'aria-labelledby': ariaLabelledby }))
        }

        if (ariaInvalid) {
            extensions.push(EditorView.contentAttributes.of({ 'aria-invalid': ariaInvalid }))
        }

        if (ariaBusy) {
            extensions.push(EditorView.contentAttributes.of({ 'aria-busy': ariaBusy }))
        }

        if (tabIndex !== 0) {
            extensions.push(EditorView.contentAttributes.of({ tabIndex: tabIndex.toString() }))
        }

        if (preventNewLine) {
            // NOTE: If a submit handler is assigned to the query input then the pressing
            // enter won't insert a line break anyway. In that case, this extensions ensures
            // that line breaks are stripped from pasted input.
            extensions.push(singleLine)
        } else {
            // Automatically enable line wrapping in multi-line mode
            extensions.push(EditorView.lineWrapping)
        }

        if (placeholder) {
            // Passing a DOM element instead of a string makes the CodeMirror
            // extension set aria-hidden="true" on the placeholder, which is
            // what we want.
            const element = document.createElement('span')
            element.append(document.createTextNode(placeholder))
            extensions.push(placeholderExtension(element))
        }

        if (editorOptions?.readOnly) {
            extensions.push(EditorView.editable.of(false))
        }

        if (searchHistory) {
            extensions.push(searchHistoryFacet.of(searchHistory))
        }

        if (onSelectSearchFromHistory) {
            extensions.push(
                EditorState.transactionExtender.of(transaction => {
                    if (transaction.isUserEvent(HISTORY_USER_EVENT)) {
                        onSelectSearchFromHistory()
                    }
                    return null
                })
            )
        }
        return extensions
    }, [
        ariaLabel,
        ariaLabelledby,
        ariaInvalid,
        ariaBusy,
        tabIndex,
        autocompletion,
        placeholder,
        preventNewLine,
        editorOptions,
        searchHistory,
        onSelectSearchFromHistory,
    ])

    // Update callback functions via effects. This avoids reconfiguring the
    // whole editor when a callback changes.
    useEffect(() => {
        if (editor) {
            setCallbacks(editor, {
                onChange,
                onSubmit,
                onFocus,
                onBlur,
                onCompletionItemSelected,
            })
        }
    }, [editor, onChange, onSubmit, onFocus, onBlur, onCompletionItemSelected])

    // Always focus the editor on 'selectedSearchContextSpec' change
    useOnValueChanged(selectedSearchContextSpec, () => {
        if (selectedSearchContextSpec && selectedSearchContextSpec) {
            editorReference.current?.focus()
        }
    })

    // Focus the editor if the autoFocus prop is truthy
    useEffect(() => {
        if (editor && autoFocus) {
            editor.focus()
        }
    }, [editor, autoFocus])

    useUpdateEditorFromQueryState(editorReference, queryState, startCompletion)

    // It looks like <Shortcut ... /> needs a stable onMatch callback, hence we
    // are storing the editor in a ref so that `globalFocus` is stable.
    const globalFocus = useCallback(() => {
        if (editorReference.current && !!document.activeElement && !isInputElement(document.activeElement)) {
            editorReference.current.focus()
        }
    }, [editorReference])

    return (
        <>
            <CodeMirrorQueryInput
                onEditorCreated={editorCreated}
                patternType={patternType}
                interpretComments={interpretComments}
                value={queryState.query}
                className={className}
                extensions={extensions}
            />
            {focusSearchBarShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={globalFocus} />
            ))}
        </>
    )
}

/**
 * Helper hook to run the function whenever the provided changes
 * (but only after the initial render)
 */
function useOnValueChanged<T = unknown>(value: T, func: () => void): void {
    const previousValue = useRef(value)

    useEffect(() => {
        if (previousValue.current !== value) {
            func()
            previousValue.current = value
        }
    }, [value, func])
}

const EMPTY: any[] = []

interface CodeMirrorQueryInputProps extends SearchPatternTypeProps {
    value: string
    onEditorCreated?: (editor: EditorView) => void
    // Whether comments are parsed and highlighted
    interpretComments?: boolean
    className?: string
    extensions: Extension[]
}

/**
 * "Core" codemirror query input component. Provides the basic behavior such as
 * theming, syntax highlighting and token info.
 */
export const CodeMirrorQueryInput: React.FunctionComponent<CodeMirrorQueryInputProps> = React.memo(
    ({ onEditorCreated, patternType, interpretComments, value, className, extensions = EMPTY }) => {
        const containerRef = useRef<HTMLDivElement | null>(null)
        const isLightTheme = useIsLightTheme()
        const externalExtensions = useMemo(() => new Compartment(), [])
        const themeExtension = useMemo(() => new Compartment(), [])

        const editorRef = useRef<EditorView | null>(null)

        // Update pattern type and/or interpretComments when changed
        useEffect(() => {
            editorRef.current?.dispatch({ effects: setQueryParseOptions.of({ patternType, interpretComments }) })
        }, [patternType, interpretComments])

        // Update theme if it changes
        useEffect(() => {
            editorRef.current?.dispatch({
                effects: themeExtension.reconfigure(EditorView.darkTheme.of(isLightTheme === false)),
            })
        }, [themeExtension, isLightTheme])

        // Update external extensions if they changed
        useEffect(() => {
            editorRef.current?.dispatch({ effects: externalExtensions.reconfigure(extensions) })
        }, [externalExtensions, extensions])

        useCodeMirror(
            editorRef,
            containerRef,
            value,
            useMemo(
                () => [
                    keymap.of(historyKeymap),
                    keymap.of(defaultKeymap),
                    history(),
                    themeExtension.of(EditorView.darkTheme.of(isLightTheme === false)),
                    parseInputAsQuery({ patternType, interpretComments }),
                    queryDiagnostic(),
                    // The precedence of these extensions needs to be decreased
                    // explicitly, otherwise the diagnostic indicators will be
                    // hidden behind the highlight background color
                    Prec.low([tokenInfo(), querySyntaxHighlighting, decorateActiveFilter, filterPlaceholder]),
                    externalExtensions.of(extensions),
                ],
                // patternType and interpretComments are updated via a
                // transaction since there is no need to re-initialize all
                // extensions
                // The extensions passed in via `extensions` are update via a
                // compartment
                // The theme (`isLightTheme`) is also updated via a compartment
                // eslint-disable-next-line react-hooks/exhaustive-deps
                [themeExtension, externalExtensions]
            )
        )

        // Notify parent component about editor instance. Among other things,
        // having a reference to the editor allows other components to initiate
        // transactions.
        useEffect(() => {
            if (editorRef.current) {
                onEditorCreated?.(editorRef.current)
            }
        }, [onEditorCreated])

        return (
            <TraceSpanProvider name="CodeMirrorQueryInput">
                <div
                    ref={containerRef}
                    className={classNames(styles.root, className, 'test-query-input', 'test-editor')}
                    data-editor="codemirror6"
                />
            </TraceSpanProvider>
        )
    }
)
CodeMirrorQueryInput.displayName = 'CodeMirrorQueryInput'

/**
 * Update the editor's value, selection and cursor depending on how the search
 * query was changed.
 */
export function useUpdateEditorFromQueryState(
    editorRef: RefObject<EditorView | null>,
    queryState: QueryState,
    startCompletion: (view: EditorView) => void
): void {
    const startCompletionRef = useRef(startCompletion)

    useEffect(() => {
        startCompletionRef.current = startCompletion
    }, [startCompletion])

    useEffect(() => {
        const editor = editorRef.current
        if (!editor) {
            return
        }

        if (queryState.changeSource === QueryChangeSource.userInput) {
            // Don't react to user input
            return
        }

        const changes =
            editor.state.sliceDoc() !== queryState.query
                ? { from: 0, to: editor.state.doc.length, insert: queryState.query }
                : undefined
        editor.dispatch({
            // Update value if it's different
            changes,
            selection: queryState.selectionRange
                ? // Select the specified range (most of the time this will be a
                  // placeholder filter value).
                  EditorSelection.range(queryState.selectionRange.start, queryState.selectionRange.end)
                : // Place the cursor at the end of the query if it changed.
                changes
                ? EditorSelection.cursor(queryState.query.length)
                : undefined,
            scrollIntoView: true,
        })

        if (queryState.hint) {
            if ((queryState.hint & EditorHint.Focus) === EditorHint.Focus) {
                editor.focus()
            }
            if ((queryState.hint & EditorHint.ShowSuggestions) === EditorHint.ShowSuggestions) {
                startCompletionRef.current(editor)
            }
            if ((queryState.hint & EditorHint.Blur) === EditorHint.Blur) {
                editor.contentDOM.blur()
            }
        }
    }, [editorRef, queryState])
}

// The remainder of the file defines all the extensions that provide the query
// editor behavior. Here is also a brief overview over CodeMirror's architecture
// to make more sense of this (see https://codemirror.net/6/docs/guide/ for more
// details):
//
// In its own words, CodeMirror has a "Functional Core" and an "Imperative
// Shell". Updates to the editor's state are performed via transactions and
// produce a new state. This new state is passed to the editor's view, which in
// turn updates itself according to the new state.
//
// Almost all behavior is implemented via extensions. There are various types of
// extensions and sometimes multiple different types are needed to implement
// some functionality:
//
// - Transaction filters can inspect transactions to change them or remove them.
// - StateFields store arbitrary values, often some form of configuration. The
// value of a state field can be updated via transactions. The fields can be
// accessed anywhere where there is access to the field object and to the
// editor's state.
// - StateEffects are a way to update a StateField's value. A StateField can
// inspect a transaction to see whether it has a specific effect and extact the
// value from the effect.
// This implementation uses this to update the query parsing options that are
// passed to the editor from the outer component.
// - Facets provide external extension points and allow aggregation of multiple
// inputs into a single output value. They also make it possible to derive
// values from the editor's current state. A concrete example is the built-in
// `EditorView.decorations` facet. A new extension can be created by calling the
// facets `compute` method, which will compute a new list of decorations based
// on the editor's current state.
// This is how the query syntax highlighting and diagnostic decorations are
// implemented.
// As another example, the diagnostics information is also stored in a facet.
// We could eventually have multiple sources that compute diagnostics
// information (which gets combined into a single list). The extension which
// computes the diagnostics decorations doesn't (need to) know where all the
// information came from.
// - ViewPlugins provide a way to extend the editor's UI in various ways. They
// can access the editor's current state and update decorations or DOM elements
// based on it.
//
// Sometimes it's not always obvious which type of extension to use to achieve a
// certain goal (and I don't claim that the implementation below is optimal).

// Instead of deriving extensions directly form props, these event handlers are
// configured via a field. This means that their values can be updated via
// transactions instead of having to reconfigure the whole editor. This is
// especially useful if the event handlers are not stable across re-renders.
// Instead of creating a separate field for every handler, all handlers are set
// via a single field to keep complexity manageable.
const [callbacksField, setCallbacks] = createUpdateableField<
    Pick<CodeMirrorQueryInputFacadeProps, 'onChange' | 'onSubmit' | 'onFocus' | 'onBlur' | 'onCompletionItemSelected'>
>({ onChange: () => {} }, callbacks => [
    Prec.high(
        keymap.of([
            {
                key: 'Enter',
                run: view => {
                    const { onSubmit } = view.state.field(callbacks)
                    if (onSubmit) {
                        // Cancel/close any open completion popovers
                        closeCompletion(view)
                        onSubmit()
                        return true
                    }
                    return false
                },
            },
        ])
    ),
    EditorView.updateListener.of((update: ViewUpdate) => {
        const { state, view } = update
        const { onChange, onFocus, onBlur, onCompletionItemSelected } = state.field(callbacks)

        if (update.docChanged) {
            onChange({
                query: state.sliceDoc(),
                changeSource: QueryChangeSource.userInput,
            })
        }

        // The focus and blur event handlers are implemented via state update handlers
        // because it appears that binding them as DOM event handlers triggers them at
        // the moment they are bound if the editor is already in that state ((not)
        // focused). See https://github.com/sourcegraph/sourcegraph/issues/37721#issuecomment-1166300433
        if (update.focusChanged) {
            if (view.hasFocus) {
                onFocus?.()
            } else {
                closeCompletion(view)
                onBlur?.()
            }
        }
        if (
            onCompletionItemSelected &&
            update.transactions.some(transaction => transaction.isUserEvent('input.complete'))
        ) {
            onCompletionItemSelected()
        }
    }),
])
