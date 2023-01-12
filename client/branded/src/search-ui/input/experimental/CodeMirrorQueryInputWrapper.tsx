import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { defaultKeymap, historyKeymap, history as codemirrorHistory } from '@codemirror/commands'
import { Compartment, EditorState, Extension, Prec } from '@codemirror/state'
import { EditorView, keymap, placeholder as placeholderExtension } from '@codemirror/view'
import { mdiClose } from '@mdi/js'
import classNames from 'classnames'
import { History } from 'history'
import { useHistory } from 'react-router'
import useResizeObserver from 'use-resize-observer'
import * as uuid from 'uuid'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { Shortcut } from '@sourcegraph/shared/src/react-shortcuts'
import { QueryChangeSource, QueryState } from '@sourcegraph/shared/src/search'
import { Icon } from '@sourcegraph/wildcard'

import { singleLine } from '../codemirror'
import { parseInputAsQuery } from '../codemirror/parsedQuery'
import { filterHighlight, querySyntaxHighlighting } from '../codemirror/syntax-highlighting'

import { editorConfigFacet, Source, suggestions } from './suggestionsExtension'

import styles from './CodeMirrorQueryInputWrapper.module.scss'

interface ExtensionConfig {
    popoverID: string
    patternType: SearchPatternType
    interpretComments: boolean
    isLightTheme: boolean
    placeholder: string
    onChange: (querySate: QueryState) => void
    onSubmit?: () => void
    suggestionsContainer: HTMLDivElement | null
    suggestionSource?: Source
    history: History
}

// For simplicity we will recompute all extensions when input changes using
// this ocmpartment
const extensionsCompartment = new Compartment()

// Helper function to update extensions dependent on props. Used when
// creating the editor and to update it when the props change.
function configureExtensions({
    popoverID,
    patternType,
    interpretComments,
    isLightTheme,
    placeholder,
    onChange,
    onSubmit,
    suggestionsContainer,
    suggestionSource,
    history,
}: ExtensionConfig): Extension {
    const extensions = [
        singleLine,
        EditorView.darkTheme.of(isLightTheme === false),
        parseInputAsQuery({ patternType, interpretComments }),
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
        // Passing a DOM element instead of a string makes the CodeMirror
        // extension set aria-hidden="true" on the placeholder, which is
        // what we want.
        const element = document.createElement('span')
        element.append(document.createTextNode(placeholder))
        extensions.push(placeholderExtension(element))
    }

    if (onSubmit) {
        extensions.push(
            editorConfigFacet.of({ onSubmit }),
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
        extensions.push(suggestions(popoverID, suggestionsContainer, suggestionSource, history))
    }

    return extensions
}

function createEditor(
    parent: HTMLDivElement,
    popoverID: string,
    queryState: QueryState,
    extensions: Extension
): EditorView {
    return new EditorView({
        state: EditorState.create({
            doc: queryState.query,
            extensions: [
                EditorView.lineWrapping,
                EditorView.contentAttributes.of({
                    role: 'combobox',
                    'aria-controls': popoverID,
                    'aria-owns': popoverID,
                    'aria-haspopup': 'grid',
                }),
                keymap.of(historyKeymap),
                keymap.of(defaultKeymap),
                codemirrorHistory(),
                Prec.low([querySyntaxHighlighting, filterHighlight]),
                EditorView.theme({
                    '&': {
                        flex: 1,
                        backgroundColor: 'var(--input-bg)',
                        borderRadius: 'var(--border-radius)',
                        borderColor: 'var(--border-color)',
                    },
                    '&.cm-editor.cm-focused': {
                        outline: 'none',
                    },
                    '.cm-content': {
                        caretColor: 'var(--search-query-text-color)',
                        fontFamily: 'var(--code-font-family)',
                        fontSize: 'var(--code-font-size)',
                        color: 'var(--search-query-text-color)',
                    },
                }),
                extensionsCompartment.of(extensions),
            ],
        }),
        parent,
    })
}

function updateEditor(editor: EditorView | null, extensions: Extension): void {
    if (editor) {
        editor.dispatch({ effects: extensionsCompartment.reconfigure(extensions) })
    }
}

function updateValueIfNecessary(editor: EditorView | null, queryState: QueryState): void {
    if (editor && queryState.changeSource !== QueryChangeSource.userInput) {
        editor.dispatch({ changes: { from: 0, to: editor.state.doc.length, insert: queryState.query } })
    }
}

export interface CodeMirrorQueryInputWrapperProps {
    queryState: QueryState
    onChange: (queryState: QueryState) => void
    onSubmit: () => void
    isLightTheme: boolean
    interpretComments: boolean
    patternType: SearchPatternType
    placeholder: string
    suggestionSource: Source
    history: History
}

export const CodeMirrorQueryInputWrapper: React.FunctionComponent<CodeMirrorQueryInputWrapperProps> = ({
    queryState,
    onChange,
    onSubmit,
    isLightTheme,
    interpretComments,
    patternType,
    placeholder,
    suggestionSource,
}) => {
    const history = useHistory()
    const [container, setContainer] = useState<HTMLDivElement | null>(null)
    const focusContainerRef = useRef<HTMLDivElement | null>(null)
    const [suggestionsContainer, setSuggestionsContainer] = useState<HTMLDivElement | null>(null)
    const popoverID = useMemo(() => uuid.v4(), [])

    // Wraps the onSubmit prop because that one changes whenever the input
    // value changes causing unnecessary reconfiguration of the extensions
    const onSubmitRef = useRef(onSubmit)
    onSubmitRef.current = onSubmit
    const hasSubmitHandler = !!onSubmit

    // Update extensions whenever any of these props change
    const extensions = useMemo(
        () =>
            configureExtensions({
                popoverID,
                patternType,
                interpretComments,
                isLightTheme,
                placeholder,
                onChange,
                onSubmit: hasSubmitHandler ? (): void => onSubmitRef.current?.() : undefined,
                suggestionsContainer,
                suggestionSource,
                history,
            }),
        [
            popoverID,
            patternType,
            interpretComments,
            isLightTheme,
            placeholder,
            onChange,
            hasSubmitHandler,
            onSubmitRef,
            suggestionsContainer,
            suggestionSource,
            history,
        ]
    )

    const editor = useMemo(
        () => (container ? createEditor(container, popoverID, queryState, extensions) : null),
        // Should only run once when the component is created, not when
        // extensions for state update (this is handled in separate hooks)
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [container]
    )
    const editorRef = useRef(editor)
    editorRef.current = editor
    useEffect(() => () => editor?.destroy(), [editor])

    // Update editor content whenever query state changes
    useEffect(() => updateValueIfNecessary(editorRef.current, queryState), [queryState])

    // Update editor configuration whenever extensions change
    useEffect(() => updateEditor(editorRef.current, extensions), [extensions])

    const focus = useCallback(() => {
        editorRef.current?.contentDOM.focus()
    }, [editorRef])

    const clear = useCallback(() => {
        onChange({ query: '' })
    }, [onChange])

    const { ref: spacerRef, height: spacerHeight } = useResizeObserver({
        ref: focusContainerRef,
    })

    const hasValue = queryState.query.length > 0

    return (
        <div className={styles.container}>
            {/* eslint-disable-next-line react/forbid-dom-props */}
            <div className={styles.spacer} style={{ height: `${spacerHeight}px` }} />
            <div className={styles.root}>
                <div ref={spacerRef} className={styles.focusContainer}>
                    <div ref={setContainer} className="d-contents" />
                    <button
                        type="button"
                        className={classNames({ [styles.showWhenFocused]: hasValue })}
                        onClick={clear}
                    >
                        <Icon svgPath={mdiClose} aria-label="Clear" />
                    </button>
                    <button
                        type="button"
                        className={classNames(styles.globalShortcut, styles.hideWhenFocused)}
                        onClick={focus}
                    >
                        /
                    </button>
                </div>
                <div ref={setSuggestionsContainer} className={styles.suggestions} />
            </div>
            <Shortcut ordered={['/']} onMatch={focus} />
        </div>
    )
}
