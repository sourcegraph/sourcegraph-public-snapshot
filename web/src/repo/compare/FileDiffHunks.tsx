import { findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { NEVER, Subject, Subscription } from 'rxjs'
import { ExtensionsProps } from '../../backend/features'
import * as GQL from '../../backend/graphqlschema'

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
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string

    hunk: GQL.IFileDiffHunk
    lineNumbers: boolean

    location: H.Location
    history: H.History
}> = ({ fileDiffAnchor, hunk, lineNumbers, location, history }) => {
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
                .map((line, i) => {
                    if (line[0] !== '+') {
                        oldLine++
                    }
                    if (line[0] !== '-') {
                        newLine++
                    }
                    const oldAnchor = `${fileDiffAnchor}L${oldLine - 1}`
                    const newAnchor = `${fileDiffAnchor}R${newLine - 1}`
                    return (
                        <tr
                            key={i}
                            className={`diff-hunk__line ${line[0] === ' ' ? 'diff-hunk__line--both' : ''} ${
                                line[0] === '-' ? 'diff-hunk__line--deletion' : ''
                            } ${line[0] === '+' ? 'diff-hunk__line--addition' : ''} ${
                                (line[0] !== '+' && location.hash === '#' + oldAnchor) ||
                                (line[0] !== '-' && location.hash === '#' + newAnchor)
                                    ? 'diff-hunk__line--active'
                                    : ''
                            }`}
                        >
                            {lineNumbers && (
                                <>
                                    {line[0] !== '+' ? (
                                        <td
                                            className="diff-hunk__num"
                                            data-line={oldLine - 1}
                                            data-part="old"
                                            id={oldAnchor}
                                            // tslint:disable-next-line:jsx-no-lambda need access to props
                                            onClick={() => history.push({ hash: oldAnchor })}
                                        />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}
                                    {line[0] !== '-' ? (
                                        <td
                                            className="diff-hunk__num"
                                            data-line={newLine - 1}
                                            data-part="new"
                                            id={newAnchor}
                                            // tslint:disable-next-line:jsx-no-lambda need access to props
                                            onClick={() => history.push({ hash: newAnchor })}
                                        />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}
                                </>
                            )}
                            <td className="diff-hunk__content">{line}</td>
                        </tr>
                    )
                })}
        </>
    )
}

interface Part {
    repoPath: string
    repoID: GQL.ID
    rev: string
    commitID: string
    filePath: string | null
}

interface Props extends ExtensionsProps {
    /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
    fileDiffAnchor: string

    /** The base repository, revision, and file. */
    base: Part

    /** The head repository, revision, and file. */
    head: Part

    /** The file's hunks. */
    hunks: GQL.IFileDiffHunk[]

    /** Whether to show line numbers. */
    lineNumbers: boolean

    className: string
    location: H.Location
    history: H.History
    hoverifier: Hoverifier
}

interface State {}

/** Displays hunks in a unified file diff. */
export class FileDiffHunks extends React.Component<Props, State> {
    /** Emits whenever the ref callback for the code element is called */
    private codeElements = new Subject<HTMLElement | null>()
    private nextCodeElement = (element: HTMLElement | null) => this.codeElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (element: HTMLElement | null) => this.blobElements.next(element)

    /** Emits whenever something is hovered in the code */
    private codeMouseOvers = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeMouseOver = (event: React.MouseEvent<HTMLElement>) => this.codeMouseOvers.next(event)

    /** Emits whenever something is hovered in the code */
    private codeMouseMoves = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeMouseMove = (event: React.MouseEvent<HTMLElement>) => this.codeMouseMoves.next(event)

    /** Emits whenever something is clicked in the code */
    private codeClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeClick = (event: React.MouseEvent<HTMLElement>) => {
        event.persist()
        this.codeClicks.next(event)
    }

    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<void>()

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: Props) {
        super(props)
        this.state = {
            hoverOverlayIsFixed: false,
            clickedGoToDefinition: false,
            mouseIsMoving: false,
        }

        this.subscriptions.add(
            this.props.hoverifier.hoverify({
                positionEvents: this.codeElements.pipe(findPositionsFromEvents()),
                positionJumps: NEVER, // TODO support diff URLs
                resolveContext: hoveredToken => {
                    const { repoPath, rev, filePath, commitID } = this.props[
                        // if part is undefined, it doesn't matter whether we chose head or base, the line stayed the same
                        hoveredToken.part === 'old' ? 'base' : 'head'
                    ]
                    // If a hover or go-to-definition was invoked on this part, we know the file path must exist
                    return { repoPath, filePath: filePath!, rev, commitID }
                },
            })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next()
    }

    public shouldComponentUpdate(nextProps: Readonly<Props>, nextState: Readonly<State>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`file-diff-hunks ${this.props.className}`} ref={this.nextBlobElement}>
                {this.props.hunks.length === 0 ? (
                    <div className="text-muted m-2">No changes</div>
                ) : (
                    <div
                        className="file-diff-hunks__container"
                        ref={this.nextCodeElement}
                        onMouseOver={this.nextCodeMouseOver}
                        onMouseMove={this.nextCodeMouseMove}
                        onClick={this.nextCodeClick}
                    >
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
                                    <DiffHunk
                                        key={i}
                                        hunk={hunk}
                                        fileDiffAnchor={this.props.fileDiffAnchor}
                                        lineNumbers={this.props.lineNumbers}
                                        location={this.props.location}
                                        history={this.props.history}
                                    />
                                ))}
                            </tbody>
                        </table>
                    </div>
                )}
            </div>
        )
    }
}
