import React, { forwardRef, useCallback, useMemo, useRef } from 'react'

import { closeCompletion, startCompletion } from '@codemirror/autocomplete'
import { type Extension, Prec, EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'

import { useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { useKeyboardShortcut } from '@sourcegraph/shared/src/keyboardShortcuts/useKeyboardShortcut'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { QueryChangeSource } from '@sourcegraph/shared/src/search'
import { appendContextFilter } from '@sourcegraph/shared/src/search/query/transformer'
import { fetchStreamSuggestions as defaultFetchStreamSuggestions } from '@sourcegraph/shared/src/search/suggestions'
import type { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'

import { BaseCodeMirrorQueryInput, type BaseCodeMirrorQueryInputProps } from './BaseCodeMirrorQueryInput'
import { createDefaultSuggestions, placeholder as placeholderExtension } from './codemirror'
import { decorateActiveFilter, filterPlaceholder } from './codemirror/active-filter'
import { queryDiagnostic } from './codemirror/diagnostics'
import { HISTORY_USER_EVENT, searchHistory as searchHistoryFacet } from './codemirror/history'
import { useMutableValue, useOnValueChanged, useUpdateInputFromQueryState } from './codemirror/react'
import { tokenInfo } from './codemirror/token-info'
import type { QueryInputProps } from './QueryInput'

import styles from './CodeMirrorQueryInput.module.scss'

export interface CodeMirrorQueryInputFacadeProps extends QueryInputProps {
    readOnly?: boolean

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
    readOnly,
    ariaLabel = 'Search query',
    ariaLabelledby,
    ariaInvalid,
    ariaBusy,
    tabIndex = 0,
    searchHistory,
    onSelectSearchFromHistory,
    // Used by the VSCode extension (which doesn't use this component directly,
    // but added for future compatibility)
    fetchStreamSuggestions = defaultFetchStreamSuggestions,
}) => {
    const editorRef = useRef<EditorView | null>(null)
    const focusSearchBarShortcut = useKeyboardShortcut('focusSearch')
    const navigate = useNavigate()
    const selectedSearchContextSpecRef = useMutableValue(selectedSearchContextSpec)

    const setEditor = useCallback(
        (editor: EditorView) => {
            editorRef.current = editor
            onEditorCreated?.(editor)
        },
        [onEditorCreated]
    )

    // Handlers
    const onSubmitRef = useMutableValue(onSubmit)
    const onChangeRef = useMutableValue(onChange)
    const onBlurRef = useMutableValue(onBlur)

    const onSubmitHandler = useCallback(
        (view: EditorView): boolean => {
            if (onSubmitRef.current) {
                // Cancel/close any open completion popovers
                closeCompletion(view)
                onSubmitRef.current()
                return true
            }
            return false
        },
        [onSubmitRef]
    )
    const onChangeHandler = useCallback(
        (value: string) => onChangeRef.current?.({ query: value, changeSource: QueryChangeSource.userInput }),
        [onChangeRef]
    )
    const onBlurHandler = useCallback(
        (view: EditorView): void => {
            // Cancel/close any open completion popovers
            closeCompletion(view)
            onBlurRef.current?.()
        },
        [onBlurRef]
    )

    const autocompletion = useCompartment(
        editorRef,
        useMemo(
            () =>
                createDefaultSuggestions({
                    fetchSuggestions: query =>
                        fetchStreamSuggestions(appendContextFilter(query, selectedSearchContextSpecRef.current)),
                    isSourcegraphDotCom,
                    navigate,
                }),
            [isSourcegraphDotCom, navigate, fetchStreamSuggestions, selectedSearchContextSpecRef]
        )
    )

    const dynamicExtensions = useCompartment(
        editorRef,
        useMemo(() => {
            const extensions: Extension[] = []
            const attributes: Record<string, string> = {}

            if (ariaLabel) {
                attributes['aria-label'] = ariaLabel
            }

            if (ariaLabelledby) {
                attributes['aria-labelledby'] = ariaLabelledby
            }

            if (ariaInvalid) {
                attributes['aria-invalid'] = ariaInvalid
            }

            if (ariaBusy) {
                attributes['aria-busy'] = ariaBusy
            }

            if (tabIndex !== 0) {
                attributes['tab-index'] = tabIndex.toString()
            }

            extensions.push(EditorView.contentAttributes.of(attributes))

            if (placeholder) {
                extensions.push(placeholderExtension(placeholder))
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
            placeholder,
            searchHistory,
            onSelectSearchFromHistory,
        ])
    )

    const extensions = useMemo(() => [autocompletion, dynamicExtensions], [autocompletion, dynamicExtensions])

    // Always focus the editor on 'selectedSearchContextSpec' change
    useOnValueChanged(selectedSearchContextSpec, () => {
        if (selectedSearchContextSpec) {
            editorRef.current?.focus()
        }
    })

    useUpdateInputFromQueryState(editorRef, queryState, startCompletion)

    // It looks like <Shortcut ... /> needs a stable onMatch callback, hence we
    // are storing the editor in a ref so that `globalFocus` is stable.
    const globalFocus = useCallback(() => {
        if (editorRef.current && !!document.activeElement && !isInputElement(document.activeElement)) {
            editorRef.current.focus()
        }
    }, [editorRef])

    return (
        <>
            <CodeMirrorQueryInput
                ref={setEditor}
                patternType={patternType}
                interpretComments={interpretComments ?? false}
                value={queryState.query}
                className={className}
                extension={extensions}
                onEnter={onSubmitHandler}
                onChange={onChangeHandler}
                onBlur={onBlurHandler}
                onFocus={onFocus}
                readOnly={readOnly}
                multiLine={!preventNewLine}
                autoFocus={autoFocus}
            />
            {focusSearchBarShortcut?.keybindings.map((keybinding, index) => (
                <Shortcut key={index} {...keybinding} onMatch={globalFocus} />
            ))}
        </>
    )
}

interface CodeMirrorQueryInputProps extends BaseCodeMirrorQueryInputProps {}

const staticExtension: Extension = [
    queryDiagnostic(),
    // The precedence of these extensions needs to be decreased
    // explicitly, otherwise the diagnostic indicators will be
    // hidden behind the highlight background color
    Prec.low([tokenInfo(), decorateActiveFilter, filterPlaceholder]),
]

/**
 * Simple query input which offers query diagnostics, token hover information,
 * and active filter highlight.
 */
export const CodeMirrorQueryInput = forwardRef<EditorView, CodeMirrorQueryInputProps>(function CodeMirrorQueryInput(
    { extension, className, ...props },
    ref
) {
    return (
        <BaseCodeMirrorQueryInput
            ref={ref}
            {...props}
            className={classNames(className, styles.root)}
            extension={useMemo(() => [extension ?? [], staticExtension], [extension])}
        />
    )
})
