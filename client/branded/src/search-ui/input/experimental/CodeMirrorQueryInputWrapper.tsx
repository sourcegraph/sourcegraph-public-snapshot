import {
    type FC,
    forwardRef,
    type MutableRefObject,
    type PropsWithChildren,
    useCallback,
    useEffect,
    useId,
    useMemo,
    useRef,
    useState,
    useImperativeHandle,
} from 'react'

import { EditorSelection, EditorState, type Extension, Prec } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import classNames from 'classnames'
import { useNavigate } from 'react-router-dom'
import useResizeObserver from 'use-resize-observer'

import { type Editor, useCompartment, viewToEditor } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { BaseCodeMirrorQueryInput } from '../BaseCodeMirrorQueryInput'
import { queryDiagnostic } from '../codemirror/diagnostics'
import { placeholder as placeholderExtension } from '../codemirror/placeholder'
import { useUpdateInputFromQueryState } from '../codemirror/react'
import { tokenInfo } from '../codemirror/token-info'

import { overrideContextOnPaste } from './codemirror/searchcontext'
import { filterDecoration } from './codemirror/syntax-highlighting'
import { modeScope, useInputMode } from './modes'
import { showWhenEmptyWithoutContext } from './placeholder'
import { Suggestions } from './Suggestions'
import {
    type Source,
    suggestions,
    startCompletion,
    getSuggestionsState,
    type Option,
    type Action,
    applyAction,
} from './suggestionsExtension'

import styles from './CodeMirrorQueryInputWrapper.module.scss'

interface ExtensionConfig {
    popoverID: string
    placeholder: string
    suggestionSource?: Source
    navigate: (destination: string) => void
}

// Helper function to update extensions dependent on props. Used when
// creating the editor and to update it when the props change.
function configureExtensions({ popoverID, placeholder, suggestionSource, navigate }: ExtensionConfig): Extension {
    const extensions = []

    if (placeholder) {
        extensions.push(placeholderExtension(placeholder, showWhenEmptyWithoutContext))
    }

    if (suggestionSource) {
        extensions.push(
            suggestions({
                id: popoverID,
                source: suggestionSource,
                navigate,
            })
        )
    }

    return extensions
}

// Creates extensions that don't depend on props
const position0 = EditorSelection.single(0)
const staticExtensions: Extension = [
    EditorState.transactionFilter.of(transaction => {
        // This is a hacky way to "fix" the cursor position when the input receives
        // focus by clicking outside of it in Chrome.
        // Debugging has revealed that in such a case the transaction has a user event
        // 'select', the new selection is set to `0` and 'scrollIntoView' is 'false'.
        // This is different from other events that change the cursor position:
        // - Clicking on text inside the input (whether focused or not) will be a 'select.pointer'
        //   user event.
        // - Moving the cursor with arrow keys will be a 'select' user event but will also set
        //   'scrollIntoView' to 'true'
        // - Entering new characters will be of user type 'input'
        // - Selecting a text range will be of user type 'select.pointer'
        // - Tabbing to the input seems to only trigger a 'select' user event transaction when
        //   the user clicked outside the input (also only in Chrome, this transaction doesn't
        //   occur in Firefox)

        if (
            !transaction.isUserEvent('select.pointer') &&
            transaction.isUserEvent('select') &&
            !transaction.scrollIntoView &&
            transaction.selection?.eq(position0)
        ) {
            return [transaction, { selection: EditorSelection.single(transaction.newDoc.length) }]
        }
        return transaction
    }),
    modeScope([queryDiagnostic(), overrideContextOnPaste], [null]),
    Prec.low([modeScope([tokenInfo(), filterDecoration], [null])]),
    EditorView.theme({
        '&': {
            flex: 1,
            // To change code mirror input area color via CSS property
            backgroundColor: 'var(--search-box-color)',
            borderRadius: 'var(--border-radius)',
            borderColor: 'var(--border-color)',
            // To ensure that the input doesn't overflow the parent
            minWidth: 0,
            marginRight: '0.5rem',
        },
        '&.cm-editor.cm-focused': {
            outline: 'none',
        },
        '.cm-scroller': {
            overflowX: 'hidden',
        },
        '.cm-content': {
            paddingLeft: '0.25rem',
        },
        '.cm-content.focus-visible': {
            boxShadow: 'none',
        },
        '.sg-decorated-token-hover': {
            borderRadius: '3px',
        },
        '.sg-query-filter-placeholder': {
            color: 'var(--text-muted)',
            fontStyle: 'italic',
        },
    }),
]

export enum QueryInputVisualMode {
    Standard = 'standard',
    Compact = 'compact',
}

export interface CodeMirrorQueryInputWrapperProps {
    queryState: QueryState
    interpretComments: boolean
    patternType: SearchPatternType
    placeholder: string
    autoFocus?: boolean
    suggestionSource?: Source
    extensions?: Extension
    visualMode?: QueryInputVisualMode | `${QueryInputVisualMode}`
    className?: string
    onChange: (queryState: QueryState) => void
    onSubmit: () => void
    onFocus?: (editor: EditorView) => void
    onBlur?: (editor: EditorView) => void
}

export const CodeMirrorQueryInputWrapper = forwardRef<Editor, PropsWithChildren<CodeMirrorQueryInputWrapperProps>>(
    (
        {
            queryState,
            interpretComments,
            patternType,
            placeholder,
            autoFocus,
            suggestionSource,
            extensions: externalExtensions,
            visualMode = QueryInputVisualMode.Standard,
            className,
            children,
            onChange,
            onSubmit,
            onFocus,
            onBlur,
        },
        ref
    ) => {
        // Global params
        const popoverID = useId()
        const navigate = useNavigate()

        // References
        const editorRef = useRef<EditorView | null>(null)
        useImperativeHandle(ref, () => viewToEditor(editorRef))

        // Local state
        const [mode, setMode, modeNotifierExtension] = useInputMode()
        const [suggestionsState, setSuggestionsState] = useState<ReturnType<typeof getSuggestionsState>>()
        // If auto-focus is enabled we do not want to show suggestions until the user
        // has interacted with the input.
        const [showSuggestions, setShowSuggestions] = useState(!autoFocus)

        // Handlers
        const onSubmitRef = useMutableValue(onSubmit)
        const onChangeRef = useMutableValue(onChange)

        const onSubmitHandler = useCallback(
            (view: EditorView): boolean => {
                if (onSubmitRef.current) {
                    onSubmitRef.current()
                    view.contentDOM.blur()
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

        // Update extensions whenever any of these props change
        const dynamicExtensions = useCompartment(
            editorRef,
            useMemo(
                () =>
                    configureExtensions({
                        popoverID,
                        placeholder,
                        suggestionSource,
                        navigate,
                    }),
                [popoverID, placeholder, suggestionSource, navigate]
            )
        )

        // Update editor state whenever query state changes
        useUpdateInputFromQueryState(editorRef, queryState, startCompletion)

        const allExtensions = useMemo(
            () => [
                externalExtensions ?? [],
                dynamicExtensions,
                modeNotifierExtension,
                EditorView.contentAttributes.of({
                    role: 'combobox',
                    // CodeMirror sets aria-multiline: true by default but it seems
                    // comboboxes are not allowed to be multiline
                    'aria-multiline': 'false',
                    'aria-controls': popoverID,
                    'aria-haspopup': 'grid',
                    'aria-label': 'Search query',
                }),
                EditorView.updateListener.of(update => {
                    setSuggestionsState(getSuggestionsState(update.state))
                    // Show suggestions after the user interacted with the input, either by
                    // moving the cursor (via keyboard or mouse) or by entering text.
                    if (
                        update.transactions.some(
                            tr => tr.isUserEvent('select') || tr.isUserEvent('input') || tr.isUserEvent('delete')
                        )
                    ) {
                        setShowSuggestions(true)
                    }
                }),
                staticExtensions,
            ],
            [popoverID, dynamicExtensions, externalExtensions, modeNotifierExtension, setShowSuggestions]
        )

        // Position cursor at the end of the input when the input changes from external sources.
        // This is necessary because the initial value might be set asynchronously.
        useEffect(() => {
            const editor = editorRef.current
            if (editor && queryState.changeSource !== QueryChangeSource.userInput) {
                editor.dispatch({
                    selection: { anchor: editor.state.doc.length },
                })
            }
        }, [queryState])

        const focus = useCallback(() => {
            editorRef.current?.focus()
        }, [])

        const toggleHistoryMode = useCallback(() => {
            if (editorRef.current) {
                setMode(editorRef.current, mode => (mode === 'History' ? null : 'History'))
                editorRef.current.focus()
            }
        }, [setMode])

        const onSelect = useCallback(
            (option: Option, action: Action) => {
                const view = editorRef.current
                if (view) {
                    applyAction(view, action ?? option.action, option, 'mouse')
                    // This is needed because clicking on an option removes focus from the query input
                    window.requestAnimationFrame(() => view.focus())
                }
            },
            [editorRef]
        )

        const { ref: inputContainerRef, height = 0 } = useResizeObserver({ box: 'border-box' })

        return (
            <div
                ref={inputContainerRef}
                className={classNames(styles.container, className, 'test-v2-query-input', 'test-editor', {
                    [styles.containerCompact]: visualMode === QueryInputVisualMode.Compact,
                })}
                role="search"
                data-editor="v2"
            >
                <div className={styles.focusContainer}>
                    <SearchModeSwitcher mode={mode} onModeChange={toggleHistoryMode} />
                    <BaseCodeMirrorQueryInput
                        ref={editorRef}
                        className={styles.input}
                        value={queryState.query}
                        autoFocus={autoFocus}
                        patternType={patternType}
                        interpretComments={interpretComments}
                        extension={allExtensions}
                        onEnter={onSubmitHandler}
                        onChange={onChangeHandler}
                        multiLine={false}
                        onFocus={onFocus}
                        onBlur={onBlur}
                    />
                    {!mode && children}
                </div>
                <div
                    className={classNames(styles.suggestions, showSuggestions && styles.open)}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ paddingTop: height }}
                >
                    {showSuggestions && suggestionsState && (
                        <Suggestions
                            id={popoverID}
                            open={suggestionsState.open}
                            activeRowIndex={suggestionsState.selectedOption}
                            results={suggestionsState.result.groups}
                            onSelect={onSelect}
                        />
                    )}
                </div>
                <Shortcut ordered={['/']} onMatch={focus} />
            </div>
        )
    }
)
CodeMirrorQueryInputWrapper.displayName = 'CodeMirrorInputWrapper'

interface SearchModeSwitcherProps {
    mode: string | null
    className?: string
    onModeChange: () => void
}

const SearchModeSwitcher: FC<SearchModeSwitcherProps> = props => {
    const { mode, className, onModeChange } = props

    return (
        <div className={classNames(className, styles.mode, !!mode && styles.modeActive)}>
            <Tooltip content="Recent searches">
                <Button variant="icon" aria-label="Open search history" onClick={onModeChange}>
                    <Icon svgPath={mdiClockOutline} aria-hidden="true" />
                    {mode && <span className="ml-1">{mode}:</span>}
                </Button>
            </Tooltip>
        </div>
    )
}

function useMutableValue<T>(value: T): MutableRefObject<T> {
    const valueRef = useRef(value)

    useEffect(() => {
        valueRef.current = value
    }, [value])

    return valueRef
}
