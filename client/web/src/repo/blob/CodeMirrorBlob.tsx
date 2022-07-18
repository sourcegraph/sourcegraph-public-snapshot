/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Compartment, Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { useHistory, useLocation } from 'react-router'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { editorHeight, useCodeMirror } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { BlobProps, updateBrowserHistoryIfNecessary } from './Blob'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './CodeMirrorLineNumbers'

const staticExtensions: Extension = [
    EditorView.editable.of(false),
    editorHeight({ height: '100%' }),
    EditorView.theme({
        '&': {
            fontFamily: 'var(--code-font-family)',
            fontSize: 'var(--code-font-size)',
            backgroundColor: 'var(--code-bg)',
        },
        '.selected-line': {
            backgroundColor: 'var(--code-selection-bg)',
        },
    }),
]

export const Blob: React.FunctionComponent<BlobProps> = ({ className, blobInfo, wrapCode, isLightTheme }) => {
    const [container, setContainer] = useState<HTMLDivElement | null>(null)

    const [dynamicExtensions, updateExtensions] = useExtension((wrapCode: boolean, isLightTheme: boolean) => [
        wrapCode ? EditorView.lineWrapping : [],
        EditorView.darkTheme.of(isLightTheme === false),
    ])

    const history = useHistory()
    const historyRef = useRef(history)
    historyRef.current = history

    const location = useLocation()
    const locationRef = useRef(location)
    locationRef.current = location

    const onSelection = useCallback((range: SelectedLineRange) => {
        const parameters = new URLSearchParams(locationRef.current.search)
        let query: string | undefined

        if (range?.line !== range?.endLine && range?.endLine) {
            query = toPositionOrRangeQueryParameter({
                range: {
                    start: { line: range.line },
                    end: { line: range.endLine },
                },
            })
        } else if (range?.line) {
            query = toPositionOrRangeQueryParameter({ position: { line: range.line } })
        }

        updateBrowserHistoryIfNecessary(
            historyRef.current,
            locationRef.current,
            addLineRangeQueryParameter(parameters, query)
        )
    }, [])

    const extensions = useMemo(() => [staticExtensions, dynamicExtensions, selectableLineNumbers({ onSelection })], [
        dynamicExtensions,
        onSelection,
    ])

    const editor = useCodeMirror(container, blobInfo.content, extensions)

    // Update extensions when prop change
    useEffect(() => {
        if (editor) {
            updateExtensions(editor, wrapCode, isLightTheme)
        }
    }, [editor, updateExtensions, wrapCode, isLightTheme])

    // Update selected lines when URL changes
    const position = useMemo(() => parseQueryAndHash(location.search, location.hash), [location.search, location.hash])
    useEffect(() => {
        if (editor) {
            selectLines(editor, position.line ? position : null)
        }
    }, [editor, position])

    return <div ref={setContainer} className={`${className} overflow-hidden`} />
}

/**
 * Helper hook for creating an extension that depends on on some input props.
 * If this proves to be useful it should be moved to the shared CodeMirror code.
 *
 * The provided callback is not called by this hook directly. It is only called
 * when the returned setter is called. It's the responsibily of the component to
 * call the setter with all necessary values.
 *
 * Like with useState, the callback function is ignored in subsequent renders,
 * so it shouldn't close over any variables that change during render.
 */
export function useExtension<T extends unknown[]>(
    callback: (...args: T) => Extension
): [Extension, (editor: EditorView, ...args: T) => void] {
    return useMemo(() => {
        const compartment = new Compartment()
        return [
            compartment.of([]),
            (editor, ...args) => editor.dispatch({ effects: compartment.reconfigure(callback(...args)) }),
        ]
        // callback is intentionally ignored in subsequent renders
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}
