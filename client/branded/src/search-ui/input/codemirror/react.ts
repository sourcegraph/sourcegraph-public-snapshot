// Hooks to provide convenience wrappers around CodeMirror extensions

import { type MutableRefObject, type RefObject, useEffect, useRef } from 'react'

import type { EditorView } from '@codemirror/view'

import type { QueryState } from '@sourcegraph/shared/src/search'

import { updateFromQueryState } from './queryState'

/**
 * Update the editor's value, selection and cursor depending on how the search
 * query was changed.
 */
export function useUpdateInputFromQueryState(
    editorRef: RefObject<EditorView>,
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
        updateFromQueryState(editor, queryState, { startCompletion: startCompletionRef.current })
    }, [editorRef, queryState])
}

/**
 * Helper function creating a ref and updating it when it changes.
 */
export function useMutableValue<T>(value: T): MutableRefObject<T> {
    const valueRef = useRef(value)

    useEffect(() => {
        valueRef.current = value
    }, [value])

    return valueRef
}

/**
 * Helper hook to run the function whenever the provided changes. Unlike
 * useEffect this hook _doesn't_ execute the callback on first render.
 */
export function useOnValueChanged<T = unknown>(value: T, func: () => void): void {
    const previousValue = useRef(value)

    useEffect(() => {
        if (previousValue.current !== value) {
            func()
            previousValue.current = value
        }
    }, [value, func])
}
