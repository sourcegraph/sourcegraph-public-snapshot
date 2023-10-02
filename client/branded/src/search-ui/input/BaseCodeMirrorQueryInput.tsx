import { type RefObject, forwardRef, memo, useEffect, useImperativeHandle, useMemo, useRef } from 'react'

import { defaultKeymap, historyKeymap, history } from '@codemirror/commands'
import type { Extension } from '@codemirror/state'
import { EditorView, drawSelection, keymap } from '@codemirror/view'
import classNames from 'classnames'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { type Editor, useCodeMirror, useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import { parseInputAsQuery, setQueryParseOptions } from './codemirror/parsedQuery'
import { querySyntaxHighlighting } from './codemirror/syntax-highlighting'

const EMPTY: Extension = []

// These extensions do not depend on any component props
const staticExtensions: Extension = [
    // Default keybindings to ensure the editor behaves correctly
    keymap.of(defaultKeymap),
    // Additional keybindings history support
    keymap.of(historyKeymap),
    // History support. It allows the user, together with the history keybindings
    // to redo/undo input changes.
    history(),
    // Let CodeMirror style selected text to make it work better with decorations
    // that change the background color.
    drawSelection(),
    // Apply syntax highlighting to query elements
    querySyntaxHighlighting,
    // The input is styled deliberately without a border so that it
    // can be integrated with various other UI elements.
    EditorView.theme({
        '.cm-content': {
            caretColor: 'var(--search-query-text-color)',
            color: 'var(--search-query-text-color)',
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            // Reset default padding
            padding: 0,
            // We need 1px padding to make the cursor visible at position 0
            paddingLeft: '1px',
        },
        '.cm-line': {
            // Reset default padding
            padding: 0,
        },
        '&.cm-focused .cm-selectionLayer .cm-selectionBackground': {
            backgroundColor: 'var(--code-selection-bg-2)',
        },
        '.cm-selectionLayer .cm-selectionBackground': {
            backgroundColor: 'var(--code-selection-bg)',
        },
    }),
]

export interface BaseCodeMirrorQueryInputProps {
    className?: string

    /**
     * The current value of the input
     */
    value: string

    /**
     * Additional CodeMirror extensions.
     */
    extension?: Extension

    /**
     * Called when the editor was instantiated.
     */
    onEditorCreated?: (editor: EditorView) => void

    // Query specific props
    /**
     * Pattern type used to parse the query. This influences
     * syntax highlighting.
     */
    patternType: SearchPatternType

    /**
     * Whether or not to recognize comments in the query input.
     */
    interpretComments: boolean
}

/**
 * BaseCodeMirrorQueryInput is a minimal query input. It provides syntax highlighting
 * and base theming.
 * Any additional functionality (autocompletion, diagnostics, token info, etc) has to be provided via extensions.
 */
export const BaseCodeMirrorQueryInput = memo(
    forwardRef<Editor, BaseCodeMirrorQueryInputProps>(
        ({ onEditorCreated, patternType, interpretComments, value, className, extension = EMPTY }, ref) => {
            const containerRef = useRef<HTMLDivElement | null>(null)
            const editorRef = useRef<EditorView | null>(null)

            const parsedQueryExtension = useQueryParser(editorRef, patternType, interpretComments)
            const externalExtension = useCompartment(editorRef, extension)
            const allExtensions = useMemo(
                () => [staticExtensions, parsedQueryExtension, externalExtension],
                [parsedQueryExtension, externalExtension]
            )

            useCodeMirror(editorRef, containerRef, value, allExtensions)

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
                        className={classNames(className, 'test-query-input', 'test-editor')}
                        data-editor="codemirror6"
                    />
                </TraceSpanProvider>
            )
        }
    )
)
BaseCodeMirrorQueryInput.displayName = 'BaseCodeMirrorQueryInput'

function useQueryParser(
    editorRef: RefObject<EditorView>,
    patternType: SearchPatternType,
    interpretComments: boolean
): Extension {
    // Update pattern type and/or interpretComments when changed and update parsed representation of query input
    useEffect(() => {
        editorRef.current?.dispatch({ effects: setQueryParseOptions.of({ patternType, interpretComments }) })
    }, [editorRef, patternType, interpretComments])

    // We only need to compute this extension on first render. In subsequent renders
    // the above effect updates the parser parameters.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    return useMemo(() => parseInputAsQuery({ patternType, interpretComments }), [])
}
