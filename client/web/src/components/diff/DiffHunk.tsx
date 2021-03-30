/* eslint jsx-a11y/click-events-have-key-events: warn, jsx-a11y/no-noninteractive-element-interactions: warn */
import * as H from 'history'
import * as React from 'react'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
    decorationStyleForTheme,
} from '../../../../shared/src/api/extension/api/decorations'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { property, isDefined } from '../../../../shared/src/util/types'
import { ThemeProps } from '../../../../shared/src/theme'
import { FileDiffHunkFields, DiffHunkLineType } from '../../graphql-operations'
import { groupBy, mapValues } from 'lodash'
interface DiffBoundaryProps extends FileDiffHunkFields {
    lineNumberClassName: string
    contentClassName: string
    lineNumbers: boolean
}

const diffHunkTypeIndicators: Record<DiffHunkLineType, string> = {
    ADDED: '+',
    UNCHANGED: ' ',
    DELETED: '-',
}

const DiffBoundary: React.FunctionComponent<DiffBoundaryProps> = props => (
    <tr className="diff-boundary">
        {props.lineNumbers && <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={0} />}
        <td className={`diff-boundary__content ${props.contentClassName}`} data-diff-marker=" ">
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <code>
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </code>
            )}
        </td>
    </tr>
)

interface Line extends FileDiffHunkFields {
    highlight: {
        aborted: boolean
        lines: Array<{
            kind: DiffHunkLineType
            html: string
            line: number
            oldLine: number
            newLine: number
        }>
    }
}

interface LineCounter {
    line: number
    dataPart: string
    className?: string
}

interface DiffHunkProps extends ThemeProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string
    hunk: Line
    lineNumbers: boolean
    decorations: Record<'head' | 'base', DecorationMapByLine>
    location: H.Location
    history: H.History
    /**
     * Reflect selected line in url
     *
     * @default true
     */
    persistLines?: boolean
}

const LineCounter: React.FunctionComponent<LineCounter> = ({ line, dataPart, className }) => {
    return (
        <td
            className={`diff-hunk__num ${className}`}
            data-line={line}
            data-part={dataPart}
            // id={oldAnchor}
            // onClick={() => persistLines && history.push({ hash: oldAnchor })}
        />
    )
}

const LineContent: React.FunctionComponent<any> = ({
    kind,
    decorations,
    line,
    isLightTheme,
    html,
    className,
}: {
    decorations: any
    kind: DiffHunkLineType
    html: string
    line: number
    isLightTheme: any
    className?: string
}) => {
    const decorationsForLine = [
        // If the line was deleted, look for decorations in the base revision
        ...((kind === DiffHunkLineType.DELETED && decorations.base.get(line)) || []),
        // If the line wasn't deleted, look for decorations in the head revision
        ...((kind !== DiffHunkLineType.DELETED && decorations.head.get(line)) || []),
    ]

    const lineStyle = decorationsForLine
        .filter(decoration => decoration.isWholeLine)
        .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
        .reduce((style, decoration) => ({ ...style, ...decoration }), {})

    return (
        <td
            className={`diff-hunk__content ${className}`}
            /* eslint-disable-next-line react/forbid-dom-props */
            style={lineStyle}
            data-diff-marker={diffHunkTypeIndicators[kind]}
        >
            <div className="d-inline-block" dangerouslySetInnerHTML={{ __html: html }} />
            {decorationsForLine.filter(property('after', isDefined)).map((decoration, index) => {
                const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)
                return (
                    <React.Fragment key={index}>
                        {' '}
                        <LinkOrSpan
                            to={decoration.after.linkURL}
                            data-tooltip={decoration.after.hoverMessage}
                            style={style}
                        >
                            {decoration.after.contentText}
                        </LinkOrSpan>
                    </React.Fragment>
                )
            })}
        </td>
    )
}

const LineGroup: React.FunctionComponent<any> = ({ decorations, isLightTheme, line, html, dataPart, className }) => {
    return (
        <>
            <LineCounter line={line} dataPart={dataPart} className={className} />
            <LineContent
                className={className}
                decorations={decorations}
                isLightTheme={isLightTheme}
                line={line}
                html={html}
            />
        </>
    )
}

export const DiffHunk: React.FunctionComponent<DiffHunkProps> = ({
    fileDiffAnchor,
    decorations,
    hunk,
    lineNumbers,
    location,
    history,
    persistLines = true,
    isLightTheme,
}) => {
    let oldLine = hunk.oldRange.startLine
    let newLine = hunk.newRange.startLine

    const lines = hunk.highlight.lines.map(line => {
        if (line.kind !== DiffHunkLineType.ADDED) {
            oldLine++
            line.line = oldLine - 1
            line.oldLine = oldLine - 1
        }
        if (line.kind !== DiffHunkLineType.DELETED) {
            newLine++
            line.line = newLine - 1
            line.newLine = newLine - 1
        }
        return line
    })

    const grouped = mapValues(groupBy(lines, 'line'))

    const result = Object.keys(grouped).map(x => {
        // if (grouped[x].length === 1 && grouped[x][0].kind === DiffHunkLineType.ADDED) {}
        return (
            <tr>
                {grouped[x].length === 1 && grouped[x][0].kind === DiffHunkLineType.ADDED ? (
                    <>
                        <td />
                        <td />
                        <LineGroup
                            className="diff-hunk__line--addition"
                            dataPart="head"
                            key={`${grouped[x][0].newLine}${grouped[x][0].html}`}
                            line={grouped[x][0].newLine}
                            html={grouped[x][0].html}
                            decorations={decorations}
                            isLightTheme={isLightTheme}
                        />
                    </>
                ) : (
                    grouped[x].map((l, index) => {
                        // const oldAnchor = `${fileDiffAnchor}L${oldLine - 1}`
                        // const newAnchor = `${fileDiffAnchor}R${newLine - 1}`
                        if (l.kind === DiffHunkLineType.UNCHANGED) {
                            return (
                                <React.Fragment key={`${l.newLine}${l.html}`}>
                                    <LineGroup
                                        dataPart="base"
                                        line={l.newLine}
                                        html={l.html}
                                        decorations={decorations}
                                        isLightTheme={isLightTheme}
                                    />
                                    <LineGroup
                                        dataPart="head"
                                        line={l.oldLine}
                                        html={l.html}
                                        decorations={decorations}
                                        isLightTheme={isLightTheme}
                                    />
                                </React.Fragment>
                            )
                        } else if (l.kind === DiffHunkLineType.DELETED) {
                            return (
                                <LineGroup
                                    className="diff-hunk__line--deletion"
                                    dataPart="base"
                                    key={`${l.oldLine}${l.html}`}
                                    line={l.oldLine}
                                    html={l.html}
                                    decorations={decorations}
                                    isLightTheme={isLightTheme}
                                />
                            )
                        } else if (l.kind === DiffHunkLineType.ADDED) {
                            return (
                                <LineGroup
                                    className="diff-hunk__line--addition"
                                    dataPart="head"
                                    key={`${l.newLine}${l.html}`}
                                    line={l.newLine}
                                    html={l.html}
                                    decorations={decorations}
                                    isLightTheme={isLightTheme}
                                />
                            )
                        }
                    })
                )}
            </tr>
        )
    })

    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
            {result}
        </>
    )
}
