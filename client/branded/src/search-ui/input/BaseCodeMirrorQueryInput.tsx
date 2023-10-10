import { type RefObject, forwardRef, memo, useEffect, useMemo, useRef } from 'react'

import { defaultKeymap, historyKeymap, history } from '@codemirror/commands'
import type { Extension } from '@codemirror/state'
import { EditorView, drawSelection, keymap } from '@codemirror/view'
import classNames from 'classnames'
import { useMergeRefs } from 'use-callback-ref'

import { TraceSpanProvider } from '@sourcegraph/observability-client'
import { useCodeMirror, useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import type { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'

import {
    type QueryInputEventHandlers,
    searchInputEventHandlers,
    setSearchInputEventHandlers,
} from './codemirror/event-handlers'
import { multiline, toSingleLine } from './codemirror/multiline'
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

export interface BaseCodeMirrorQueryInputProps extends QueryInputEventHandlers {
    // General element props

    className?: string

    // Common input props

    /**
     * The current value of the input
     */
    value: string

    /**
     * If enabled (default: false) the input value cannot be edited by the user.
     * (the value can still be changed via the {@link value} prop or via
     * CodeMirror extensions).
     */
    readOnly?: boolean

    /**
     * If enabled (default: false) the input will receive focus when it mounts.
     * Changing the value afterwards won't have any effect.
     */
    autoFocus?: boolean

    // Query input specific props

    /**
     * If enabled (default: false) supports multi-line input. The input will grow vertically
     * as necessary and line wrapping is turned on. If `onEnter` is set it will only
     * be triggered on `Mod-Enter`.
     * If not set all line breaks will be replaced by a space instead and line wrapping
     * is disabled.
     */
    multiLine?: boolean

    /**
     * Pattern type used to parse the query. This influences syntax highlighting.
     */
    patternType: SearchPatternType

    /**
     * When set the query parser will be instructed to consider comments.
     */
    interpretComments: boolean

    // CodeMirror specific props

    /**
     * Additional CodeMirror extensions.
     * These extensions are placed _before_ the component's own extensions so that they
     * have a higher priority by default. This component's own extensions all have
     * "default" priority.
     */
    extension?: Extension
}

/**
 * BaseCodeMirrorQueryInput is a minimal query input. It provides syntax highlighting
 * and base theming.
 * Any additional functionality (autocompletion, diagnostics, token info, etc) has to be provided via extensions.
 */
export const BaseCodeMirrorQueryInput = memo(
    forwardRef<EditorView, BaseCodeMirrorQueryInputProps>(function BaseCodeMirrorQueryInput(
        {
            className,
            value,
            readOnly = false,
            multiLine = false,
            autoFocus = false,
            patternType,
            interpretComments,
            extension = EMPTY,
            onChange,
            onEnter,
            onFocus,
            onBlur,
        },
        ref
    ) {
        const containerRef = useRef<HTMLDivElement | null>(null)
        const localEditorRef = useRef<EditorView | null>(null)
        const editorRef = useMergeRefs([ref, localEditorRef])
        // We need to expliclity remove line breaks when multiLine is set, to ensure
        // that the initial value is properly formatted. The corresponding CodeMirror
        // extension only affects changes made after initialization.
        const normalizedValue = useMemo(() => (multiLine ? value : toSingleLine(value)), [multiLine, value])

        const eventHandlers = useEventHandlers(
            editorRef,
            useMemo(
                () => ({
                    onEnter,
                    onChange,
                    onFocus,
                    onBlur,
                }),
                [onEnter, onChange, onFocus, onBlur]
            )
        )
        const parsedQueryExtension = useQueryParser(editorRef, patternType, interpretComments)
        const readOnlyExtension = useCompartment(
            editorRef,
            useMemo(() => EditorView.editable.of(!readOnly), [readOnly])
        )
        const multiLineExtension = useCompartment(
            editorRef,
            useMemo(() => multiline(multiLine), [multiLine])
        )
        const externalExtension = useCompartment(editorRef, extension)
        const allExtensions = useMemo(
            () => [
                externalExtension,
                parsedQueryExtension,
                eventHandlers,
                multiLineExtension,
                readOnlyExtension,
                staticExtensions,
            ],
            [eventHandlers, parsedQueryExtension, readOnlyExtension, multiLineExtension, externalExtension]
        )

        useCodeMirror(editorRef, containerRef, normalizedValue, allExtensions)

        // useEffects after this point can access the editor instance via editorRef.current during the initial render

        useEffect(() => {
            if (autoFocus) {
                // NOTE: editorRef.current will be set at this point
                editorRef.current?.focus()
            }
            // autFocus should only have an effect on first render
            // eslint-disable-next-line react-hooks/exhaustive-deps
        }, [])

        return (
            <TraceSpanProvider name="CodeMirrorQueryInput">
                <div
                    ref={containerRef}
                    className={classNames(className, 'test-query-input', 'test-editor')}
                    data-editor="codemirror6"
                />
            </TraceSpanProvider>
        )
    })
)

function useQueryParser(
    editorRef: RefObject<EditorView>,
    patternType: SearchPatternType,
    interpretComments: boolean
): Extension {
    const shouldUpdate = useRef(false)

    // Update pattern type and/or interpretComments when changed and update parsed representation of query input
    useEffect(() => {
        if (shouldUpdate.current) {
            editorRef.current?.dispatch({ effects: setQueryParseOptions.of({ patternType, interpretComments }) })
        } else {
            shouldUpdate.current = true
        }
    }, [editorRef, patternType, interpretComments])

    // We only need to compute this extension on first render. In subsequent renders
    // the above effect updates the parser parameters.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    return useMemo(() => parseInputAsQuery({ patternType, interpretComments }), [])
}

/**
 * Initializes the query input handlers on first render and updates them via state effects
 * on updates.
 */
function useEventHandlers(editorRef: RefObject<EditorView>, handlers: QueryInputEventHandlers): Extension {
    const shouldUpdate = useRef(false)
    useEffect(() => {
        if (shouldUpdate.current && editorRef.current) {
            setSearchInputEventHandlers(editorRef.current, handlers)
        } else {
            shouldUpdate.current = true
        }
    }, [editorRef, handlers])

    // We only need to compute this extension on first render. In subsequent renders
    // the above effect updates the handlers.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    return useMemo(() => searchInputEventHandlers.init(() => handlers), [])
}
