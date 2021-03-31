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
import { useSplitDiff } from '@sourcegraph/wildcard/src/hooks'
import { useHooksAddLineNumber } from '@sourcegraph/wildcard/src/hooks'

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
    const { hunksWithLineNumber } = useHooksAddLineNumber(
        hunk.highlight.lines,
        hunk.newRange.startLine,
        hunk.oldRange.startLine
    )
    const { diff } = useSplitDiff(hunksWithLineNumber)

    const singleHunk = () => {
        return diff.map(elements => {
            return (
                <tr>
                    {elements.map((line, index) => {
                        if (!line)
                            return (
                                <>
                                    <td />
                                    <td />
                                </>
                            )
                        if (line.kind === DiffHunkLineType.ADDED) {
                            return (
                                <LineGroup
                                    className="diff-hunk__line--addition"
                                    dataPart="head"
                                    key={`${line.newLine}${line.html}`}
                                    line={line.newLine}
                                    html={line.html}
                                    decorations={decorations}
                                    isLightTheme={isLightTheme}
                                />
                            )
                        }

                        if (line.kind === DiffHunkLineType.DELETED) {
                            return (
                                <LineGroup
                                    className="diff-hunk__line--deletion"
                                    dataPart="base"
                                    key={`${line.newLine}${line.html}`}
                                    line={line.oldLine}
                                    html={line.html}
                                    decorations={decorations}
                                    isLightTheme={isLightTheme}
                                />
                            )
                        }

                        if (line.kind === DiffHunkLineType.UNCHANGED) {
                            let l = elements[index + 1] ? line.oldLine : line.newLine
                            return (
                                <LineGroup
                                    className="diff-hunk__line"
                                    dataPart="base"
                                    key={`${line.newLine}${line.html}`}
                                    line={l}
                                    html={line.html}
                                    decorations={decorations}
                                    isLightTheme={isLightTheme}
                                />
                            )
                        } else {
                            return null
                        }
                    })}
                </tr>
            )
        })
    }

    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
            {singleHunk()}
        </>
    )
}
