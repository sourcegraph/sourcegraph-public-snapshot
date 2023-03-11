import {
    forwardRef,
    PropsWithChildren,
    useCallback,
    useEffect,
    useImperativeHandle,
    useMemo,
    useRef,
    useState,
} from 'react'

import { defaultKeymap, historyKeymap, history as codemirrorHistory } from '@codemirror/commands'
import { Compartment, EditorState, Extension, Prec } from '@codemirror/state'
import { EditorView, keymap, drawSelection } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import classNames from 'classnames'
import inRange from 'lodash/inRange'
import { useNavigate } from 'react-router-dom'
import useResizeObserver from 'use-resize-observer'
import * as uuid from 'uuid'

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

// For simplicity we will recompute all extensions when input changes using
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
        EditorView.darkTheme.of(isLightTheme === false),
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
function createStaticExtensions({ popoverID }: { popoverID: string }): Extension {
    return [
        singleLine,
        drawSelection(),
        EditorView.contentAttributes.of({
            role: 'combobox',
            'aria-controls': popoverID,
            'aria-owns': popoverID,
            'aria-haspopup': 'grid',
        }),
        keymap.of(historyKeymap),
        keymap.of(defaultKeymap),
        codemirrorHistory(),
        filterPlaceholder,
        queryDiagnostic(),
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
}

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
            children,
        },
        ref
    ) => {
        const navigate = useNavigate()
        const editorContainerRef = useRef<HTMLDivElement | null>(null)
        const focusContainerRef = useRef<HTMLDivElement | null>(null)
        const [suggestionsContainer, setSuggestionsContainer] = useState<HTMLDivElement | null>(null)
        const popoverID = useMemo(() => uuid.v4(), [])
        const [mode, setMode, modeNotifierExtension] = useInputMode()

        const onSubmitRef = useRef(onSubmit)
        useEffect(() => {
            onSubmitRef.current = onSubmit
        }, [onSubmit])
        const hasSubmitHandler = !!onSubmit

        const onChangeRef = useRef(onChange)
        useEffect(() => {
            onChangeRef.current = onChange
        }, [onChange])

        const staticExtensions = useMemo(() => createStaticExtensions({ popoverID }), [popoverID])
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
            ]
        )

        // Update query extensions whenever any of these props change
        const queryExtensions = useMemo(
            () => configureQueryExtensions({ patternType, interpretComments }),
            [patternType, interpretComments]
        )

        const editorRef = useRef<EditorView | null>(null)

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
                    staticExtensions,
                    extensionsCompartment.of(dynamicExtensions),
                    querySettingsCompartment.of(queryExtensions),
                ],
                // Only set extensions during initialization. dynamicExtensions and queryExtensions
                // are updated separately.
                // eslint-disable-next-line react-hooks/exhaustive-deps
                []
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

        const { ref: spacerRef, height: spacerHeight } = useResizeObserver({
            ref: focusContainerRef,
        })

        return (
            <div className={styles.container}>
                {/* eslint-disable-next-line react/forbid-dom-props */}
                <div className={styles.spacer} style={{ height: `${spacerHeight}px` }} />
                <div className={styles.root}>
                    <div ref={spacerRef} className={styles.focusContainer}>
                        <div className={classNames(styles.modeSection, !!mode && styles.active)}>
                            <Tooltip content="Recent searches">
                                <Button variant="icon" onClick={toggleHistoryMode} aria-label="Open search history">
                                    <Icon svgPath={mdiClockOutline} aria-hidden="true" />
                                </Button>
                            </Tooltip>
                            {mode && <span className="ml-1">{mode}:</span>}
                        </div>
                        <div ref={editorContainerRef} className={styles.input} />
                        {!mode && children}
                    </div>
                    <div ref={setSuggestionsContainer} className={styles.suggestions} />
                </div>
                <Shortcut ordered={['/']} onMatch={focus} />
            </div>
        )
    }
)
