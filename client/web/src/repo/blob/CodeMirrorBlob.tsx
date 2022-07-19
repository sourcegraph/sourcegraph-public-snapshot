/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Extension, StateEffect, StateField, Text } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { useHistory, useLocation } from 'react-router'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import {
    editorHeight,
    replaceValue,
    useCodeMirror,
    useCompartment,
} from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { BlobProps, updateBrowserHistoryIfChanged } from './Blob'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './codemirror/linenumbers'
import { JsonDocument, Occurrence, SyntaxKind } from '../../lsif/lsif-typed'
import { HighlightRange, highlightRanges } from './codemirror/highlight'

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

export const Blob: React.FunctionComponent<BlobProps> = ({
    className,
    blobInfo,
    wrapCode,
    isLightTheme,
    ariaLabel,
    role,
}) => {
    const [container, setContainer] = useState<HTMLDivElement | null>(null)

    const settings = useMemo(
        () => [wrapCode ? EditorView.lineWrapping : [], EditorView.darkTheme.of(isLightTheme === false)],
        [wrapCode, isLightTheme]
    )

    const [settingsCompartment, updateSettingsCompartment] = useCompartment(settings)

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

        updateBrowserHistoryIfChanged(
            historyRef.current,
            locationRef.current,
            addLineRangeQueryParameter(parameters, query)
        )
    }, [])

    const extensions = useMemo(
        () => [
            staticExtensions,
            settingsCompartment,
            selectableLineNumbers({ onSelection }),
            syntaxHighlight(blobInfo.lsif),
        ],
        [settingsCompartment, onSelection]
    )

    const editor = useCodeMirror(container, blobInfo.content, extensions, { updateValueOnChange: false })

    useEffect(() => {
        if (editor) {
            updateSettingsCompartment(editor, settings)
        }
    }, [editor, updateSettingsCompartment, settings])

    // We don't want to trigger the transaction on the first render. Maybe there
    // is a better way to do this.
    const initialRender = useRef(true)
    useEffect(() => {
        if (editor && !initialRender.current) {
            editor.dispatch({
                changes: replaceValue(editor, blobInfo.content),
                effects: setSCIPData.of(blobInfo.lsif),
            })
        }
        initialRender.current = false
    }, [editor, blobInfo])

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

    return <div ref={setContainer} aria-label={ariaLabel} role={role} className={`${className} overflow-hidden`} />
}

const setSCIPData = StateEffect.define<string>()

/**
 * Helper extension to convert SCIP-encoded highlighting information to the
 * format expected by our syntax highlighting extenions. The SCIP data should be
 * set/updated via the `setSCIPData` effect.
 */
function syntaxHighlight(initialSCIPJSON: string): Extension {
    function parseAndSortRanges(doc: Text, json: string): HighlightRange[] {
        const ranges: HighlightRange[] = []

        for (const occurence of (JSON.parse(json) as JsonDocument).occurrences ?? []) {
            let {
                range: [startLine, startColumn, endLine, endColumn],
            } = occurence

            // If the range is in the same line, an occurence has only 3 fields
            if (endColumn === undefined) {
                endColumn = endLine
                endLine = startLine
            }

            const start = doc.line(startLine + 1)
            const end = startLine === endLine ? start : doc.line(endLine + 1)

            ranges.push([
                start.from + startColumn,
                end.from + endColumn,
                `hl-typed-${SyntaxKind[occurence.syntaxKind]}`,
            ])
        }

        return ranges.sort((a, b) => (a[0] === b[0] ? a[1] - b[1] : a[0] - b[0]))
    }

    return StateField.define<HighlightRange[]>({
        create: state => parseAndSortRanges(state.doc, initialSCIPJSON),

        update(value, transaction) {
            let newSCIPData = ''

            for (const effect of transaction.effects) {
                if (effect.is(setSCIPData)) {
                    newSCIPData = effect.value
                    break
                }
            }

            return newSCIPData ? parseAndSortRanges(transaction.newDoc, newSCIPData) : value
        },

        provide: field => highlightRanges.from(field),
    })
}
