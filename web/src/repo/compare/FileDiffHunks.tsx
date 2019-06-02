import { DiffPart, DOMFunctions, findPositionsFromEvents, Hoverifier } from '@sourcegraph/codeintellify'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { combineLatest, NEVER, Observable, of, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import {
    decorationAttachmentStyleForTheme,
    DecorationMapByLine,
    decorationStyleForTheme,
    groupDecorationsByLine,
} from '../../../../shared/src/api/client/services/decoration'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { LinkOrSpan } from '../../../../shared/src/components/LinkOrSpan'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import * as GQL from '../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { isDefined, propertyIsDefined } from '../../../../shared/src/util/types'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec, toURIWithPath } from '../../../../shared/src/util/url'
import { ThemeProps } from '../../theme'

const DiffBoundary: React.FunctionComponent<{
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
            {props.oldRange.lines !== undefined && props.newRange.lines !== undefined && (
                <code>
                    @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},
                    {props.newRange.lines} {props.section && `@@ ${props.section}`}
                </code>
            )}
        </td>
    </tr>
)

const DiffHunk: React.FunctionComponent<
    {
        /** The anchor (URL hash link) of the file diff. The component creates sub-anchors with this prefix. */
        fileDiffAnchor: string

        hunk: GQL.IFileDiffHunk
        lineNumbers: boolean
        decorations: Record<'head' | 'base', DecorationMapByLine>

        location: H.Location
        history: H.History
    } & ThemeProps
> = ({ fileDiffAnchor, decorations, hunk, lineNumbers, location, history, isLightTheme }) => {
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
                    const decorationsForLine = [
                        // If the line was deleted, look for decorations in the base rev
                        ...((line[0] === '-' && decorations.base.get(oldLine - 1)) || []),
                        // If the line wasn't deleted, look for decorations in the head rev
                        ...((line[0] !== '-' && decorations.head.get(newLine - 1)) || []),
                        // Look for decorations in both if the line existed in both
                    ]
                    const lineStyle = decorationsForLine
                        .filter(decoration => decoration.isWholeLine)
                        .map(decoration => decorationStyleForTheme(decoration, isLightTheme))
                        .reduce((style, decoration) => ({ ...style, ...decoration }), {})

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
                                    {/* Base line number (if line existed in base rev) */}
                                    {line[0] !== '+' ? (
                                        <td
                                            className="diff-hunk__num"
                                            data-line={oldLine - 1}
                                            data-part="base"
                                            id={oldAnchor}
                                            // tslint:disable-next-line:jsx-no-lambda need access to props
                                            onClick={() => history.push({ hash: oldAnchor })}
                                        />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}

                                    {/* Head line number (if line still exists in head rev) */}
                                    {line[0] !== '-' ? (
                                        <td
                                            className="diff-hunk__num"
                                            data-line={newLine - 1}
                                            data-part="head"
                                            id={newAnchor}
                                            // tslint:disable-next-line:jsx-no-lambda need access to props
                                            onClick={() => history.push({ hash: newAnchor })}
                                        />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}
                                </>
                            )}
                            {/* tslint:disable-next-line: jsx-ban-props Needed for decorations */}
                            <td className="diff-hunk__content" style={lineStyle}>
                                {line}
                                {decorationsForLine.filter(propertyIsDefined('after')).map((decoration, i) => {
                                    const style = decorationAttachmentStyleForTheme(decoration.after, isLightTheme)
                                    return (
                                        <>
                                            {' '}
                                            <LinkOrSpan
                                                key={i}
                                                to={decoration.after.linkURL}
                                                data-tooltip={decoration.after.hoverMessage}
                                                // tslint:disable-next-line: jsx-ban-props Needed for decorations
                                                style={style}
                                            >
                                                {decoration.after.contentText}
                                            </LinkOrSpan>
                                        </>
                                    )
                                })}
                            </td>
                        </tr>
                    )
                })}
        </>
    )
}

const diffDomFunctions: DOMFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        const row = target.closest('tr')
        if (!row) {
            return null
        }
        return row.cells[2]
    },

    getCodeElementFromLineNumber: (
        codeView: HTMLElement,
        line: number,
        part?: DiffPart
    ): HTMLTableCellElement | null => {
        // For unchanged lines, prefer line number in head
        const lineNumberCell = codeView.querySelector(`[data-line="${line}"][data-part="${part || 'head'}"]`)
        if (!lineNumberCell) {
            return null
        }
        const row = lineNumberCell.parentElement as HTMLTableRowElement
        const codeCell = row.cells[2]
        return codeCell
    },

    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const [baseLineNumberCell, headLineNumberCell] = row.cells
        // For unchanged lines, prefer line number in head
        if (headLineNumberCell.dataset.line) {
            return +headLineNumberCell.dataset.line
        }
        if (baseLineNumberCell.dataset.line) {
            return +baseLineNumberCell.dataset.line
        }
        throw new Error('Neither head or base line number cell have data-line set')
    },

    getDiffCodePart: (codeCell: HTMLElement): DiffPart => {
        const row = codeCell.parentElement as HTMLTableRowElement
        const [baseLineNumberCell, headLineNumberCell] = row.cells
        if (baseLineNumberCell.dataset.part && headLineNumberCell.dataset.part) {
            return null
        }
        if (baseLineNumberCell.dataset.part) {
            return 'base'
        }
        if (headLineNumberCell.dataset.part) {
            return 'head'
        }
        throw new Error('Could not figure out diff part for code element')
    },

    isFirstCharacterDiffIndicator: (codeElement: HTMLElement) => true,
}

interface Part {
    repoName: string
    repoID: GQL.ID
    rev: string
    commitID: string

    /**
     * `null` if the file does not exist in this diff part.
     */
    filePath: string | null
}

interface FileHunksProps extends PlatformContextProps, ExtensionsControllerProps, ThemeProps {
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
    hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>
}

interface FileDiffHunksState {
    /**
     * Decorations for the file at the two revisions of the diff
     */
    decorations: Record<'head' | 'base', DecorationMapByLine>
}

/** Displays hunks in a unified file diff. */
export class FileDiffHunks extends React.Component<FileHunksProps, FileDiffHunksState> {
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
    private componentUpdates = new Subject<FileHunksProps>()

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: FileHunksProps) {
        super(props)
        this.state = {
            decorations: { head: new Map(), base: new Map() },
        }

        this.subscriptions.add(
            this.props.hoverifier.hoverify({
                dom: diffDomFunctions,
                positionEvents: this.codeElements.pipe(
                    filter(isDefined),
                    findPositionsFromEvents(diffDomFunctions)
                ),
                positionJumps: NEVER, // TODO support diff URLs
                resolveContext: hoveredToken => {
                    // if part is undefined, it doesn't matter whether we chose head or base, the line stayed the same
                    const { repoName, rev, filePath, commitID } = this.props[hoveredToken.part || 'head']
                    // If a hover or go-to-definition was invoked on this part, we know the file path must exist
                    return { repoName, filePath: filePath!, rev, commitID }
                },
            })
        )

        // Listen to decorations from extensions and group them by line
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(({ head, base, extensionsController }) => ({ head, base, extensionsController })),
                    distinctUntilChanged(
                        (a, b) =>
                            isEqual(a.head, b.head) &&
                            isEqual(a.base, b.base) &&
                            a.extensionsController !== b.extensionsController
                    ),
                    switchMap(({ head, base, extensionsController }) => {
                        const getDecorationsForPart = ({
                            repoName,
                            commitID,
                            filePath,
                        }: Part): Observable<TextDocumentDecoration[] | null> =>
                            filePath !== null
                                ? extensionsController.services.textDocumentDecoration.getDecorations({
                                      uri: toURIWithPath({ repoName, commitID, filePath }),
                                  })
                                : of(null)
                        return combineLatest([getDecorationsForPart(head), getDecorationsForPart(base)])
                    })
                )
                .subscribe(([headDecorations, baseDecorations]) => {
                    this.setState({
                        decorations: {
                            head: groupDecorationsByLine(headDecorations),
                            base: groupDecorationsByLine(baseDecorations),
                        },
                    })
                })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate(
        nextProps: Readonly<FileHunksProps>,
        nextState: Readonly<FileDiffHunksState>
    ): boolean {
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
                                        {...this.props}
                                        key={i}
                                        hunk={hunk}
                                        decorations={this.state.decorations}
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
