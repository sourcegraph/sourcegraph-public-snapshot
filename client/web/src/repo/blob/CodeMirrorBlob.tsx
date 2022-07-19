/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Extension, RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView, ViewPlugin } from '@codemirror/view'
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

    const syntaxHighlighting = useMemo(() => {
        // When CodeMirror is enabled, blobInfo.html contains a JSON blob
        // encoding ranges as [start, end, CSS class] (unless there is no syntax
        // highlighting)
        if (blobInfo.html.includes('<table>')) {
            return []
        }
        const builder = new RangeSetBuilder<Decoration>()
        const decorations: Record<string, Decoration> = {}

        for (const [start, end, spec] of parseAndSortRangesJSON(blobInfo.html)) {
            // Creating a single CSS class string (and thus a single span)
            // appears to be more performant than creating a separate decoration
            // for each class
            // TODO: Consider using a ViewPlugin to only create decorations for
            // lines that are rendered
            const cls = spec
                .split('.')
                .map(cls => `hl-${cls}`)
                .sort()
                .join(' ')
            builder.add(start, end, decorations[cls] || (decorations[cls] = Decoration.mark({ class: cls })))
        }
        // This is implemented as a ViewPlugin because of how CodeMirror gets
        // updated atm: The value and the syntax decorations are updated in two
        // separate transactions (value first then syntax highlighting). This
        // causes CodeMirror to throw an error because a lot of decorations will
        // be out of range.
        // We therefore clear the decorations when the document changes.
        // The base hook needs to be refactored so that multiple parts of the
        // editor's state can be updated at the same time when the value changes
        return ViewPlugin.define(
            () => ({
                decorations: builder.finish(),
                update(update) {
                    if (update.docChanged) {
                        // False positive?
                        // eslint-disable-next-line react/no-this-in-sfc
                        this.decorations = Decoration.none
                    }
                },
            }),
            { decorations: value => value.decorations }
        )
    }, [blobInfo.html])

    const settings = useMemo(
        () => [wrapCode ? EditorView.lineWrapping : [], EditorView.darkTheme.of(isLightTheme === false)],
        [wrapCode, isLightTheme]
    )

    const [settingsCompartment, updateSettingsCompartment] = useCompartment(settings)
    const [syntaxHighlightingCompartment, updateSyntaxHighlightingCompartment] = useCompartment(syntaxHighlighting)

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

    const extensions = useMemo(
        () => [
            staticExtensions,
            settingsCompartment,
            syntaxHighlightingCompartment,
            selectableLineNumbers({ onSelection }),
        ],
        [settingsCompartment, syntaxHighlightingCompartment, onSelection]
    )

    const editor = useCodeMirror(container, blobInfo.content, extensions)

    useEffect(() => {
        if (editor) {
            updateSettingsCompartment(editor, settings)
        }
    }, [editor, updateSettingsCompartment, settings])

    useEffect(() => {
        if (editor) {
            updateSyntaxHighlightingCompartment(editor, syntaxHighlighting)
        }
    }, [editor, updateSyntaxHighlightingCompartment, syntaxHighlighting])

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

function parseAndSortRangesJSON(ranges: string): [number, number, string][] {
    return (JSON.parse(ranges) as [number, number, string][]).sort(([startA], [startB]) => startA - startB)
}
