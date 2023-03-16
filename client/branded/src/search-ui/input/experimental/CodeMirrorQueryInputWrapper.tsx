import {
    FC,
    forwardRef,
    MutableRefObject,
    PropsWithChildren,
    useCallback,
    useEffect,
    useId,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from 'react'

import { defaultKeymap, historyKeymap, history as codemirrorHistory } from '@codemirror/commands'
import { Compartment, EditorSelection, EditorState, Extension, Prec } from '@codemirror/state'
import { EditorView, keymap, drawSelection } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import classNames from 'classnames'
import inRange from 'lodash/inRange'
import { useNavigate } from 'react-router-dom'
import useResizeObserver from 'use-resize-observer'

import { HistoryOrNavigate } from '@sourcegraph/common'
import { Editor, useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { QueryChangeSource, QueryState } from '@sourcegraph/shared/src/search'
import { getTokenLength } from '@sourcegraph/shared/src/search/query/utils'
import { Button, Icon, Tooltip } from '@sourcegraph/wildcard'

import { singleLine, placeholder as placeholderExtension } from '../codemirror'
import { filterPlaceholder } from '../codemirror/active-filter'
import { queryDiagnostic } from '../codemirror/diagnostics'
import { parseInputAsQuery, tokens } from '../codemirror/parsedQuery'
import { querySyntaxHighlighting } from '../codemirror/syntax-highlighting'
import { tokenInfo } from '../codemirror/token-info'
import { useUpdateEditorFromQueryState } from '../CodeMirrorQueryInput'

import { overrideContextOnPaste, shiftPasteOverwrite } from './codemirror/searchcontext'
import { filterDecoration } from './codemirror/syntax-highlighting'
import { modeScope, useInputMode } from './modes'
import { Source, suggestions, startCompletion } from './suggestionsExtension'

import styles from './CodeMirrorQueryInputWrapper.module.scss'

interface ExtensionConfig {
    popoverID: string
    isLightTheme: boolean
    placeholder: string
    onChange: (querySate: QueryState) => void
    onSubmit?: () => void
    suggestionsContainer: HTMLDivElement | null
    suggestionSource?: Source
    historyOrNavigate: HistoryOrNavigate
}

// We want to show a placeholder also if the query only contains a context
// filter.
function showWhenEmptyWithoutContext(state: EditorState): boolean {
    // Show placeholder when empty
    if (state.doc.length === 0) {
        return true
    }

    const queryTokens = tokens(state)

    if (queryTokens.length > 2) {
        return false
    }
    // Only show the placeholder if the cursor is at the end of the content
    if (state.selection.main.from !== state.doc.length) {
        return false
    }

    // If there are two tokens, only show the placeholder if the second one is a
    // whitespace of length 1
    if (queryTokens.length === 2 && (queryTokens[1].type !== 'whitespace' || getTokenLength(queryTokens[1]) !== 1)) {
        return false
    }

    return (
        queryTokens.length > 0 &&
        queryTokens[0].type === 'filter' &&
        queryTokens[0].field.value === 'context' &&
        !inRange(state.selection.main.from, queryTokens[0].range.start, queryTokens[0].range.end + 1)
    )
}

// For simplicity, we will recompute all extensions when input changes using
// this compartment
const extensionsCompartment = new Compartment()

// Helper function to update extensions dependent on props. Used when
// creating the editor and to update it when the props change.
function configureExtensions({
    popoverID,
    isLightTheme,
    placeholder,
    onChange,
    onSubmit,
    suggestionsContainer,
    suggestionSource,
    historyOrNavigate,
}: ExtensionConfig): Extension {
    const extensions = [
        EditorView.darkTheme.of(!isLightTheme),
        EditorView.updateListener.of(update => {
            if (update.docChanged) {
                onChange({
                    query: update.state.sliceDoc(),
                    changeSource: QueryChangeSource.userInput,
                })
            }
        }),
    ]

    if (placeholder) {
        extensions.push(placeholderExtension(placeholder, showWhenEmptyWithoutContext))
    }

    if (onSubmit) {
        extensions.push(
            Prec.high(
                keymap.of([
                    {
                        key: 'Enter',
                        run() {
                            onSubmit()
                            return true
                        },
                    },
                    {
                        key: 'Mod-Enter',
                        run() {
                            onSubmit()
                            return true
                        },
                    },
                ])
            )
        )
    }

    if (suggestionSource && suggestionsContainer) {
        extensions.push(
            suggestions({
                id: popoverID,
                parent: suggestionsContainer,
                source: suggestionSource,
                historyOrNavigate,
            })
        )
    }

    return extensions
}

// Holds extensions that somehow depend on the query or query parameters. They
// are stored in a separate compartment to avoid re-creating other extensions.
// (if we didn't do this the suggestions list would flicker because it gets
// recreated)
const querySettingsCompartment = new Compartment()

function configureQueryExtensions({
    patternType,
    interpretComments,
}: {
    patternType: SearchPatternType
    interpretComments: boolean
}): Extension {
    return parseInputAsQuery({ patternType, interpretComments })
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
    singleLine,
    drawSelection(),
    keymap.of(historyKeymap),
    keymap.of(defaultKeymap),
    codemirrorHistory(),
    filterPlaceholder,
    shiftPasteOverwrite(),
    modeScope([queryDiagnostic(), overrideContextOnPaste], [null]),
    Prec.low([querySyntaxHighlighting, modeScope([tokenInfo(), filterDecoration], [null])]),
    EditorView.theme({
        '&': {
            flex: 1,
            backgroundColor: 'var(--input-bg)',
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
            caretColor: 'var(--search-query-text-color)',
            color: 'var(--search-query-text-color)',
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            padding: 0,
            paddingLeft: '0.25rem',
        },
        '.cm-content.focus-visible': {
            boxShadow: 'none',
        },
        '.cm-line': {
            padding: 0,
        },
        '.theme-dark .cm-selectionLayer .cm-selectionBackground': {
            backgroundColor: 'var(--gray-08)',
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

function updateExtensions(editor: EditorView | null, extensions: Extension): void {
    if (editor) {
        editor.dispatch({ effects: extensionsCompartment.reconfigure(extensions) })
    }
}

function updateQueryExtensions(editor: EditorView | null, extensions: Extension): void {
    if (editor) {
        editor.dispatch({ effects: querySettingsCompartment.reconfigure(extensions) })
    }
}

const empty: any[] = []

export enum QueryInputVisualMode {
    Standard = 'standard',
    Compact = 'compact',
}

export interface CodeMirrorQueryInputWrapperProps {
    queryState: QueryState
    onChange: (queryState: QueryState) => void
    onSubmit: () => void
    isLightTheme: boolean
    interpretComments: boolean
    patternType: SearchPatternType
    placeholder: string
    suggestionSource?: Source
    extensions?: Extension
    visualMode?: QueryInputVisualMode | `${QueryInputVisualMode}`
    className?: string
}

export const CodeMirrorQueryInputWrapper = forwardRef<Editor, PropsWithChildren<CodeMirrorQueryInputWrapperProps>>(
    (
        {
            queryState,
            onChange,
            onSubmit,
            isLightTheme,
            interpretComments,
            patternType,
            placeholder,
            suggestionSource,
            extensions: externalExtensions = empty,
            visualMode = QueryInputVisualMode.Standard,
            className,
            children,
        },
        ref
    ) => {
        // Global params
        const popoverID = useId()
        const navigate = useNavigate()

        // References
        const editorRef = useRef<EditorView | null>(null)
        const editorContainerRef = useRef<HTMLDivElement | null>(null)

        // Local state
        const [mode, setMode, modeNotifierExtension] = useInputMode()
        const [suggestionsContainer, setSuggestionsContainer] = useState<HTMLDivElement | null>(null)

        // Handlers
        const onSubmitRef = useMutableValue(onSubmit)
        const onChangeRef = useMutableValue(onChange)

        const hasSubmitHandler = !!onSubmit

        // Update extensions whenever any of these props change
        const dynamicExtensions = useMemo(
            () => [
                configureExtensions({
                    popoverID,
                    isLightTheme,
                    placeholder,
                    onChange: (...args) => onChangeRef.current(...args),
                    onSubmit: hasSubmitHandler
                        ? (): void => {
                              if (onSubmitRef.current) {
                                  onSubmitRef.current()
                                  editorRef.current?.contentDOM.blur()
                              }
                          }
                        : undefined,
                    suggestionsContainer,
                    suggestionSource,
                    historyOrNavigate: navigate,
                }),
                externalExtensions,
                modeNotifierExtension,
            ],
            [
                popoverID,
                isLightTheme,
                placeholder,
                hasSubmitHandler,
                suggestionsContainer,
                suggestionSource,
                navigate,
                externalExtensions,
                modeNotifierExtension,
                onChangeRef,
                onSubmitRef,
            ]
        )

        // Update query extensions whenever any of these props change
        const queryExtensions = useMemo(
            () => configureQueryExtensions({ patternType, interpretComments }),
            [patternType, interpretComments]
        )

        // Update editor state whenever query state changes
        useUpdateEditorFromQueryState(editorRef, queryState, startCompletion)

        // Update editor configuration whenever extensions change
        useEffect(() => updateExtensions(editorRef.current, dynamicExtensions), [dynamicExtensions])
        useEffect(() => updateQueryExtensions(editorRef.current, queryExtensions), [queryExtensions])

        // Create editor
        useCodeMirror(
            editorRef,
            editorContainerRef,
            queryState.query,
            useMemo(
                () => [
                    EditorView.contentAttributes.of({
                        role: 'combobox',
                        'aria-controls': popoverID,
                        'aria-owns': popoverID,
                        'aria-haspopup': 'grid',
                    }),
                    staticExtensions,
                    extensionsCompartment.of(dynamicExtensions),
                    querySettingsCompartment.of(queryExtensions),
                ],
                // Only set extensions during initialization. dynamicExtensions and queryExtensions
                // are updated separately.
                // eslint-disable-next-line react-hooks/exhaustive-deps
                [popoverID]
            )
        )

        useImperativeHandle(
            ref,
            () => ({
                focus() {
                    editorRef.current?.focus()
                },
                blur() {
                    editorRef.current?.contentDOM.blur()
                },
            }),
            []
        )

        // Position cursor at the end of the input when it is initialized
        useEffect(() => {
            if (editorRef.current) {
                editorRef.current.dispatch({
                    selection: { anchor: editorRef.current.state.doc.length },
                })
            }
        }, [])

        const focus = useCallback(() => {
            editorRef.current?.focus()
        }, [])

        const toggleHistoryMode = useCallback(() => {
            if (editorRef.current) {
                setMode(editorRef.current, mode => (mode === 'History' ? null : 'History'))
                editorRef.current.focus()
            }
        }, [setMode])

        const { ref: inputContainerRef, height = 0 } = useResizeObserver({ box: 'border-box' })

        return (
            <div
                ref={inputContainerRef}
                className={classNames(styles.container, className, {
                    [styles.containerCompact]: visualMode === QueryInputVisualMode.Compact,
                })}
            >
                <div className={styles.focusContainer}>
                    <SearchModeSwitcher mode={mode} onModeChange={toggleHistoryMode} />
                    <div ref={editorContainerRef} className={styles.input} />
                    {!mode && children}
                </div>
                <div
                    ref={setSuggestionsContainer}
                    className={styles.suggestions}
                    // eslint-disable-next-line react/forbid-dom-props
                    style={{ paddingTop: height }}
                />
                <Shortcut ordered={['/']} onMatch={focus} />
            </div>
        )
    }
)

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
                </Button>
            </Tooltip>
            {mode && <span className="ml-1">{mode}:</span>}
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
