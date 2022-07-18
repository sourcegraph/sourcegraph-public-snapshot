/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { useHistory, useLocation } from 'react-router'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import { editorHeight, useCodeMirror, useCompartment } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
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
            // This check is necessary because at the moment the position
            // information is updated before the file content, meaning it's
            // possible that the currently loaded document has fewer lines.
            if (!position?.line || editor.state.doc.lines >= (position.endLine ?? position.line)) {
                selectLines(editor, position.line ? position : null)
            }
        }
        // blobInfo isn't used but we need to trigger the line selection and focus
        // logic whenever the content changes
    }, [editor, position, blobInfo])

    return <div ref={setContainer} className={`${className} overflow-hidden`} />
}
