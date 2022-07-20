/**
 * An experimental implementation of the Blob view using CodeMirror
 */

import { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Extension, RangeSetBuilder, StateEffect, StateField } from '@codemirror/state'
import { Decoration, DecorationSet, EditorView, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { useHistory, useLocation } from 'react-router'

import { addLineRangeQueryParameter, toPositionOrRangeQueryParameter } from '@sourcegraph/common'
import {
    editorHeight,
    replaceValue,
    useCodeMirror,
    useCompartment,
} from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'

import { BlobProps, updateBrowserHistoryIfNecessary } from './Blob'
import { selectLines, selectableLineNumbers, SelectedLineRange } from './CodeMirrorLineNumbers'

type HighlightRange = [number, number, string]

const setRanges = StateEffect.define<HighlightRange[]>()
function syntaxHighlighting(initialRange: HighlightRange[]): Extension {
    return StateField.define<HighlightRange[]>({
        create() {
            return initialRange
        },
        update(value, update) {
            for (const effect of update.effects) {
                if (effect.is(setRanges)) {
                    return effect.value
                }
            }
            return value
        },
        provide: field =>
            ViewPlugin.fromClass(
                class {
                    decorationCache: Record<string, Decoration> = {}
                    decorations: DecorationSet = Decoration.none

                    constructor(view: EditorView) {
                        this.decorations = this.computeDecorations(view)
                    }

                    update(update: ViewUpdate) {
                        if (update.docChanged) {
                            this.decorationCache = {}
                        }

                        if (
                            update.viewportChanged ||
                            update.transactions.some(transaction =>
                                transaction.effects.some(effect => effect.is(setRanges))
                            )
                        ) {
                            this.decorations = this.computeDecorations(update.view)
                        }
                    }

                    computeDecorations(view: EditorView): DecorationSet {
                        const { from, to } = view.viewport
                        const ranges = view.state.field(field)
                        const rangeIndex = rangeIndexOf(ranges, from)

                        if (rangeIndex === -1) {
                            return Decoration.none
                        }
                        const builder = new RangeSetBuilder<Decoration>()

                        for (let index = rangeIndex; index < ranges.length && ranges[index][0] <= to; index++) {
                            const [start, end, spec] = ranges[index]
                            const cls = spec
                                .split('.')
                                .map(cls => `hl-${cls}`)
                                .sort()
                                .join(' ')
                            builder.add(
                                start,
                                end,
                                this.decorationCache[cls] ||
                                    (this.decorationCache[cls] = Decoration.mark({ class: cls }))
                            )
                        }

                        return builder.finish()
                    }
                },
                { decorations: plugin => plugin.decorations }
            ),
    })
}

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
            selectableLineNumbers({ onSelection }),
            syntaxHighlighting(parseAndSortRangesJSON(blobInfo.html)),
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
                effects: setRanges.of(parseAndSortRangesJSON(blobInfo.html)),
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

    return <div ref={setContainer} className={`${className} overflow-hidden`} />
}

function parseAndSortRangesJSON(ranges: string): [number, number, string][] {
    // Workaround for current ranges implementation
    if (ranges.includes('<table>')) {
        return []
    }

    return (JSON.parse(ranges) as [number, number, string][]).sort(([startA, endA], [startB, endB]) =>
        startA === startB ? endA - endB : startA - startB
    )
}

// Performs a binary search to find the left most range whose end is start or
// the right most element whose end is < start.
// It uses the end of the range for comparison because we want all decorations
// that are applicable at a certain position.
function rangeIndexOf(ranges: [number, number, string][], start: number): number | -1 {
    let low = 0
    let high = ranges.length

    while (low < high) {
        const middle = Math.floor((low + high) / 2)
        if (ranges[middle][1] < start) {
            low = middle + 1
        } else {
            high = middle
        }
    }

    return ranges[low] === undefined ? -1 : low
}
