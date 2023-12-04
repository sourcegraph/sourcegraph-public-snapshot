import { RangeSetBuilder } from '@codemirror/state'
import {
    Decoration,
    type DecorationSet,
    type EditorView,
    type PluginValue,
    type ViewUpdate,
    ViewPlugin,
} from '@codemirror/view'
import classNames from 'classnames'

import { logger } from '@sourcegraph/common'
import { SyntaxKind } from '@sourcegraph/shared/src/codeintel/scip'

import { getLinksFromString } from '../../linkifiy/get-links'

import { syntaxHighlight } from './highlight'

import styles from './links.module.scss'

/**
 * SyntaxKinds where we will check for URLs.
 * We only match URLs within comments and string literals to reduce false positives.
 */
const SUPPORTED_KINDS = new Set<SyntaxKind>([SyntaxKind.Comment, SyntaxKind.StringLiteral])

interface LineDecoration {
    from: number
    to: number
    decoration: Decoration
}

/**
 * Iterates through `SUPPORTED_KINDS` highlighting occurrences and match any URLs within.
 * Converts matches into <a> tags with a permanent underline and a relevant href.
 */
class LinkBuilder implements PluginValue {
    public decorations: DecorationSet = Decoration.none
    private lineCache: Map<number, LineDecoration[]> = new Map()

    constructor(view: EditorView) {
        this.decorations = this.computeDecorations(view)
    }

    public update(update: ViewUpdate): void {
        if (update.viewportChanged) {
            this.decorations = this.computeDecorations(update.view)
        }
    }

    private computeDecorations(view: EditorView): DecorationSet {
        const builder = new RangeSetBuilder<Decoration>()
        try {
            const { from, to } = view.viewport

            // Determine the start and end lines of the current viewport
            const fromLine = view.state.doc.lineAt(from)
            const toLine = view.state.doc.lineAt(to)

            const { occurrences, lineIndex } = view.state.facet(syntaxHighlight)

            // Find index of first relevant token
            let startIndex: number | undefined
            {
                let line = fromLine.number - 1
                do {
                    startIndex = lineIndex[line++]
                } while (startIndex === undefined && line < lineIndex.length)
            }

            if (startIndex !== undefined) {
                // Iterate over the rendered line (numbers) and get the
                // corresponding occurrences from the highlighting table.
                const textDocument = view.state.doc

                for (let index = startIndex; index < occurrences.length; index++) {
                    const occurrence = occurrences[index]
                    const occurrenceStartLine = occurrence.range.start.line + 1

                    // Skip if out of viewport
                    if (occurrenceStartLine > toLine.number) {
                        break
                    }

                    // Skip if the syntax kind is not supported.
                    if (occurrence.kind === undefined || !SUPPORTED_KINDS.has(occurrence.kind)) {
                        continue
                    }

                    // Skip if the range is not on a single line.
                    // We know that we will not match a link here.
                    if (!occurrence.range.isSingleLine()) {
                        continue
                    }

                    const line = textDocument.line(occurrenceStartLine)

                    // Check if the line is already in the cache
                    let lineDecorations = this.lineCache.get(occurrenceStartLine)

                    if (!lineDecorations) {
                        lineDecorations = []

                        const start = line.from + occurrence.range.start.character
                        const end = line.from + occurrence.range.end.character

                        const links = getLinksFromString({
                            input: textDocument.sliceString(start, end),
                        })

                        for (const link of links) {
                            lineDecorations.push({
                                from: Math.min(start + link.start, line.to),
                                to: Math.min(start + link.end, line.to),
                                decoration: Decoration.mark({
                                    tagName: 'a',
                                    class: classNames(styles.link, `hl-typed-${SyntaxKind[occurrence.kind]}`),
                                    attributes: {
                                        href: link.href,
                                        target: '_blank',
                                        rel: 'noreferrer noopener',
                                    },
                                }),
                            })
                        }

                        // Cache the decorations for the current line
                        this.lineCache.set(occurrenceStartLine, lineDecorations)
                    }

                    // Add the decorations to the builder
                    for (const { from, to, decoration } of lineDecorations) {
                        builder.add(from, to, decoration)
                    }
                }
            }
        } catch (error) {
            logger.error('Failed to build links', error)
        }
        return builder.finish()
    }
}

/**
 * Transforms URLs within code comments and string literals into links.
 */
export const linkify = ViewPlugin.fromClass(LinkBuilder, { decorations: plugin => plugin.decorations })
