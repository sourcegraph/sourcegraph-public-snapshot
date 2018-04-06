import * as React from 'react'

const DiffBoundary: React.SFC<{
    /** The "lines" property is set for end boundaries (only for start boundaries and between hunks). */
    oldRange: { startLine: number; lines?: number }
    newRange: { startLine: number; lines?: number }

    section: string | null

    lineNumberClassName: string
    contentClassName: string

    lineNumbers: boolean
}> = props => (
    <tr className="diff-boundary">
        {props.lineNumbers && <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={2} />}
        <td className={`diff-boundary__content ${props.contentClassName}`}>
            {props.oldRange.lines !== undefined &&
                props.newRange.lines !== undefined && (
                    <code>
                        @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},{
                            props.newRange.lines
                        }{' '}
                        {props.section && `@@ ${props.section}`}
                    </code>
                )}
        </td>
    </tr>
)

const DiffHunk: React.SFC<{
    hunk: GQL.IFileDiffHunk
    lineNumbers: boolean
}> = ({ hunk, lineNumbers }) => {
    let oldLine = hunk.oldRange.startLine
    let newLine = hunk.newRange.startLine
    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
            {hunk.body
                .split('\n')
                .slice(0, -1)
                .map((line, i) => (
                    <tr
                        key={i}
                        className={`diff-hunk__line ${line[0] === ' ' ? 'diff-hunk__line--both' : ''} ${
                            line[0] === '-' ? 'diff-hunk__line--deletion' : ''
                        } ${line[0] === '+' ? 'diff-hunk__line--addition' : ''}`}
                    >
                        {lineNumbers && (
                            <>
                                {line[0] !== '+' ? (
                                    <td className="diff-hunk__num" data-line-number={oldLine++} />
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}
                                {line[0] !== '-' ? (
                                    <td className="diff-hunk__num" data-line-number={newLine++} />
                                ) : (
                                    <td className="diff-hunk__num diff-hunk__num--empty" />
                                )}
                            </>
                        )}
                        <td className="diff-hunk__content">{line}</td>
                    </tr>
                ))}
        </>
    )
}

interface Props {
    /** The file's hunks. */
    hunks: GQL.IFileDiffHunk[]

    /** Whether to show line numbers. */
    lineNumbers: boolean

    className: string
}

interface State {}

/** Displays hunks in a unified file diff. */
export class FileDiffHunks extends React.PureComponent<Props, State> {
    public render(): JSX.Element | null {
        return (
            <div className={`file-diff-hunks ${this.props.className}`}>
                <table className="file-diff-hunks__table">
                    {this.props.lineNumbers && (
                        <colgroup>
                            <col width="40" />
                            <col width="40" />
                            <col />
                        </colgroup>
                    )}
                    <tbody>
                        {this.props.hunks.map((hunk, i) => (
                            <DiffHunk key={i} hunk={hunk} lineNumbers={this.props.lineNumbers} />
                        ))}
                    </tbody>
                </table>
            </div>
        )
    }
}
