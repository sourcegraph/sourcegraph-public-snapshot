/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-noninteractive-element-interactions: warn */
import { DecorationMapByLine, decorationStyleForTheme } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isDefined, property } from '@sourcegraph/shared/src/util/types'
import * as React from 'react'
import { useLocation } from 'react-router'
import { TextDocumentDecoration } from 'sourcegraph'
import { DiffHunkLineType, FileDiffHunkFields } from '../../graphql-operations'
import { DiffBoundary } from './DiffBoundary'
import { EmptyLine, Line } from './Lines'
import { useHunksAddLineNumber } from './useHunksAddLineNumber'
import { useSplitDiff } from './useSplitDiff'

export interface DiffHunkProps extends ThemeProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string
    hunk: FileDiffHunkFields
    lineNumbers: boolean
    decorations: Record<'head' | 'base', DecorationMapByLine>
    /**
     * Reflect selected line in url
     *
     * @default true
     */
    persistLines?: boolean
}

// eslint-disable-next-line @typescript-eslint/explicit-function-return-type
const addDecorations = (isLightTheme: boolean, decorationsForLine: TextDocumentDecoration[]) => {
    const lineStyle = decorationsForLine
        .filter(decoration => decoration.isWholeLine)
        .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
        .reduce((style, decoration) => ({ ...style, ...decoration }), {})

    const decorationsWithAfterProperty = decorationsForLine.filter(property('after', isDefined))

    return { lineStyle, decorationsWithAfterProperty }
}

export interface Hunk {
    kind: DiffHunkLineType
    html: string
    anchor: string
    oldLine?: number
    newLine?: number
}

export const DiffSplitHunk: React.FunctionComponent<DiffHunkProps> = ({
    fileDiffAnchor,
    decorations,
    hunk,
    lineNumbers,
    persistLines = true,
    isLightTheme,
}) => {
    const location = useLocation()

    const { hunksWithLineNumber } = useHunksAddLineNumber(
        hunk.highlight.lines,
        hunk.newRange.startLine,
        hunk.oldRange.startLine,
        'anchor_'
    )

    const { diff } = useSplitDiff(hunksWithLineNumber)

    const groupHunks = React.useCallback(
        (hunks: Hunk[]): JSX.Element[] => {
            const elements = []
            for (let index = 0; index < hunks.length; index++) {
                const current = hunks[index]

                const lineNumber = (elements[index + 1] ? current.oldLine : current.newLine) as number
                const active = location.hash === `#${current.anchor}`

                const decorationsForLine = [
                    // If the line was deleted, look for decorations in the base revision
                    ...((current.kind === DiffHunkLineType.DELETED && decorations.base.get(lineNumber)) || []),
                    // If the line wasn't deleted, look for decorations in the head revision
                    ...((current.kind !== DiffHunkLineType.DELETED && decorations.head.get(lineNumber)) || []),
                ] as TextDocumentDecoration[]

                const { lineStyle, decorationsWithAfterProperty } = addDecorations(isLightTheme, decorationsForLine)

                if (current.kind === DiffHunkLineType.UNCHANGED) {
                    // UNCHANGED is displayed on both side
                    elements.push(
                        <tr key={current.anchor} data-testid={current.anchor}>
                            <Line
                                persistLines={persistLines}
                                lineStyle={lineStyle}
                                decorations={decorationsWithAfterProperty}
                                key={`L${current.anchor}`}
                                id={`L${current.anchor}`}
                                kind={current.kind}
                                lineNumber={current.oldLine}
                                anchor={current.anchor}
                                html={current.html}
                                className={active ? 'diff-hunk__line--active' : ''}
                                lineNumbers={lineNumbers}
                            />
                            <Line
                                persistLines={persistLines}
                                lineStyle={lineStyle}
                                decorations={decorationsWithAfterProperty}
                                key={`R${current.anchor}`}
                                id={`R${current.anchor}`}
                                kind={current.kind}
                                lineNumber={current.newLine}
                                anchor={current.anchor}
                                html={current.html}
                                className={active ? 'diff-hunk__line--active' : ''}
                                lineNumbers={lineNumbers}
                            />
                        </tr>
                    )
                } else if (current.kind === DiffHunkLineType.DELETED) {
                    const next = hunks[index + 1]
                    // If an ADDED change is following a DELETED change, they should be displayed side by side
                    if (next?.kind === DiffHunkLineType.ADDED) {
                        index = index + 1
                        elements.push(
                            <tr key={current.anchor} data-testid={current.anchor}>
                                <Line
                                    persistLines={persistLines}
                                    lineStyle={lineStyle}
                                    decorations={decorationsWithAfterProperty}
                                    key={current.anchor}
                                    kind={current.kind}
                                    lineNumber={current.oldLine}
                                    anchor={current.anchor}
                                    html={current.html}
                                    className={active ? 'diff-hunk__line--active' : ''}
                                    lineNumbers={lineNumbers}
                                />
                                <Line
                                    persistLines={persistLines}
                                    lineStyle={lineStyle}
                                    decorations={decorationsWithAfterProperty}
                                    key={next.anchor}
                                    kind={next.kind}
                                    lineNumber={next.newLine}
                                    anchor={next.anchor}
                                    html={next.html}
                                    className={location.hash === `#${next.anchor}` ? 'diff-hunk__line--active' : ''}
                                    lineNumbers={lineNumbers}
                                />
                            </tr>
                        )
                    } else {
                        // DELETED is following by an empty line
                        elements.push(
                            <tr key={current.anchor} data-testid={current.anchor}>
                                <Line
                                    persistLines={persistLines}
                                    lineStyle={lineStyle}
                                    decorations={decorationsWithAfterProperty}
                                    key={current.anchor}
                                    kind={current.kind}
                                    lineNumber={
                                        current.kind === DiffHunkLineType.DELETED ? current.oldLine : lineNumber
                                    }
                                    anchor={current.anchor}
                                    html={current.html}
                                    className={active ? 'diff-hunk__line--active' : ''}
                                    lineNumbers={lineNumbers}
                                />
                                <EmptyLine />
                            </tr>
                        )
                    }
                } else {
                    // ADDED is preceded by an empty line
                    elements.push(
                        <tr key={current.anchor} data-testid={current.anchor}>
                            <EmptyLine />
                            <Line
                                persistLines={persistLines}
                                lineStyle={lineStyle}
                                decorations={decorationsWithAfterProperty}
                                key={current.anchor}
                                kind={current.kind}
                                lineNumber={lineNumber}
                                anchor={current.anchor}
                                html={current.html}
                                className={active ? 'diff-hunk__line--active' : ''}
                                lineNumbers={lineNumbers}
                            />
                        </tr>
                    )
                }
            }

            return elements
        },
        [decorations.base, decorations.head, isLightTheme, location.hash]
    )

    const diffView = React.useMemo(() => groupHunks(diff), [diff, groupHunks])

    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
                diffMode="split"
            />
            {diffView}
        </>
    )
}
