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

    const dynamicExtensions = useMemo(
        () => [wrapCode ? EditorView.lineWrapping : [], EditorView.darkTheme.of(isLightTheme === false)],
        [wrapCode, isLightTheme]
    )

    const [compartment, updateCompartment] = useCompartment(dynamicExtensions)

    const history = useHistory()
    const location = useLocation()

    // Keep history and location in a ref so that we can use the latest value in
    // the onSelection callback without having to recreate it and having to
    // reconfigure the editor extensions
    const historyRef = useRef(history)
    historyRef.current = history
    const locationRef = useRef(location)
    locationRef.current = location

    const onSelection = useCallback((range: SelectedLineRange, event: MouseEvent) => {
        const parameters = new URLSearchParams(locationRef.current.search)
        let query: string | undefined
        // If the shift key is pressed and the URL currently contains a position
        // then we are going to replace the history entry instead of adding a
        // new one. This avoids two history entries for the common operation of
        // selecting the start line first and then selecting the end line via
        // shift+click
        const replace = event.shiftKey

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
            addLineRangeQueryParameter(parameters, query),
            replace
        )
    }, [])

    const extensions = useMemo(() => [staticExtensions, compartment, selectableLineNumbers({ onSelection })], [
        compartment,
        onSelection,
    ])

    const editor = useCodeMirror(container, blobInfo.content, extensions)

    useEffect(() => {
        if (editor) {
            updateCompartment(editor, dynamicExtensions)
        }
    }, [editor, updateCompartment, dynamicExtensions])

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
 * Helper hook for extensions that depend on on some input props.
 * With this hook the extension is isolated in a compartment so it can be
 * updated without reconfiguring the whole editor.
 *
 * If this proves to be useful it should be moved to the shared CodeMirror
 * directory.
 */
export function useCompartment(
    initialExtension: Extension
): [Extension, (editor: EditorView, extension: Extension) => void] {
    return useMemo(() => {
        const compartment = new Compartment()
        return [
            compartment.of(initialExtension),
            (editor, extension: Extension) => {
                // This check avoids an unnecessary update when the editor is
                // first created
                if (initialExtension !== extension) {
                    editor.dispatch({ effects: compartment.reconfigure(extension) })
                }
            },
        ]
        // initialExtension is intentionally ignored in subsequent renders
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])
}
