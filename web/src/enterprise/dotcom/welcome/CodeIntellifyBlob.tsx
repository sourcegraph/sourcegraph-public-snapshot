import { createHoverifier, findPositionsFromEvents, HoveredToken, HoverState } from '@sourcegraph/codeintellify'
import { getTokenAtPosition } from '@sourcegraph/codeintellify/lib/token_position'
import { Position } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { catchError, filter, map, withLatestFrom } from 'rxjs/operators'
import { ActionItemProps } from '../../../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { getHoverActions } from '../../../../../shared/src/hover/actions'
import { HoverContext, HoverOverlay } from '../../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { isDefined, propertyIsDefined } from '../../../../../shared/src/util/types'
import {
    FileSpec,
    ModeSpec,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
} from '../../../../../shared/src/util/url'
import { getHover } from '../../../backend/features'
import { fetchBlob } from '../../../repo/blob/BlobPage'

interface Props extends ExtensionsControllerProps, PlatformContextProps {
    history: H.History
    location: H.Location
    className: string
    startLine: number
    endLine: number
    parentElement: string

    tooltipClass: string
    defaultHoverPosition: Position
}

interface State extends HoverState<HoverContext, HoverMerged, ActionItemProps> {
    /**
     * The blob data or error that happened.
     * undefined while loading.
     */
    blobOrError?: GQL.IGitBlob | ErrorLike
    target?: EventTarget
}

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the decoration, return null.
        if (
            target.classList.contains('line-decoration-attachment') ||
            target.classList.contains('line-decoration-attachment__contents')
        ) {
            return null
        }

        const row = target.closest('tr')
        if (!row) {
            return null
        }

        return row.cells[1]
    },
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number): HTMLElement | null => {
        const lineNumberElement = codeView.querySelector(`td[data-line="${line}"]`)

        if (!lineNumberElement) {
            return null
        }
        return lineNumberElement.nextElementSibling as HTMLElement | null
    },
    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const numberCell = row.cells[0]
        if (!numberCell || !numberCell.dataset.line) {
            throw new Error('Could not find line number')
        }
        return parseInt(numberCell.dataset.line, 10)
    },
}

const REPO_NAME = 'github.com/gorilla/mux'
const COMMIT_ID = '9e1f5955c0d22b55d9e20d6faa28589f83b2faca'
const REV = undefined
const FILE_PATH = 'mux.go'

export class CodeIntellifyBlob extends React.Component<Props, State> {
    /** Emits whenever the ref callback for the code element is called */
    private codeViewElements = new Subject<HTMLElement | null>()
    private nextCodeViewElement = (element: HTMLElement | null) => this.codeViewElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the demo file element is called */
    private codeIntellifyBlobElements = new Subject<HTMLElement | null>()
    private nextCodeIntellifyBlobElements = (element: HTMLElement | null) =>
        this.codeIntellifyBlobElements.next(element)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent) => this.closeButtonClicks.next(event)

    private subscriptions = new Subscription()

    private componentUpdates = new Subject<void>()

    private target: EventTarget | null = null

    constructor(props: Props) {
        super(props)
        this.state = {}

        const hoverifier = createHoverifier<
            RepoSpec & RevSpec & FileSpec & ResolvedRevSpec,
            HoverMerged,
            ActionItemProps
        >({
            closeButtonClicks: this.closeButtonClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.codeIntellifyBlobElements),
                map(([, hoverOverlayElement, codeIntellifyBlobElement]) => ({
                    hoverOverlayElement,
                    codeIntellifyBlobElement,
                })),
                filter(propertyIsDefined('codeIntellifyBlobElement')),
                map(({ hoverOverlayElement, codeIntellifyBlobElement }) => ({
                    hoverOverlayElement,
                    relativeElement: codeIntellifyBlobElement.closest(this.props.parentElement) as HTMLElement | null,
                })),
                // Can't reposition HoverOverlay or file weren't rendered
                filter(propertyIsDefined('relativeElement')),
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            getHover: hoveredToken => getHover(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
            getActions: context => getHoverActions(this.props, context),
        })

        this.subscriptions.add(hoverifier)
        const positionEvents = this.codeViewElements.pipe(
            filter(isDefined),
            findPositionsFromEvents(domFunctions)
        )

        const targets = positionEvents.pipe(map(({ event: { target } }) => target))

        targets.subscribe(target => (this.target = target))

        this.subscriptions.add(
            hoverifier.hoverify({
                positionEvents,
                resolveContext: () => ({
                    repoName: REPO_NAME,
                    commitID: COMMIT_ID,
                    rev: REV || '',
                    filePath: FILE_PATH,
                }),
                dom: domFunctions,
            })
        )

        this.subscriptions.add(hoverifier.hoverStateUpdates.subscribe(update => this.setState(update)))

        this.subscriptions.add(
            this.codeViewElements
                .pipe(
                    filter(isDefined),
                    map(codeView => getTokenAtPosition(codeView, props.defaultHoverPosition, domFunctions)),
                    filter(isDefined)
                )
                .subscribe(token => {
                    const showOnHomepage = props.className === 'code-intellify-container' && window.innerWidth >= 1393
                    const showOnModal =
                        props.className === 'code-intellify-container-modal' && window.innerWidth >= 1275
                    if (showOnHomepage || showOnModal) {
                        token.click()
                    }
                })
        )
    }

    public componentDidMount(): void {
        // Fetch repository revision.
        fetchBlob({
            repoName: REPO_NAME,
            commitID: COMMIT_ID,
            filePath: FILE_PATH,
            isLightTheme: false,
            disableTimeout: false,
        })
            .pipe(
                catchError(error => {
                    console.error(error)
                    return [error]
                })
            )
            .subscribe(blobOrError => this.setState({ blobOrError }), err => console.error(err))

        this.componentUpdates.next()

        this.subscriptions.add(
            this.props.extensionsController.services.model.model.next({
                ...this.props.extensionsController.services.model.model.value,
                visibleViewComponents: [
                    {
                        type: 'textEditor',
                        item: {
                            uri: `git://github.com/gorilla/mux?9e1f5955c0d22b55d9e20d6faa28589f83b2faca#mux.go`,
                            languageId: 'go',
                            text: '',
                        },
                        selections: [],
                        isActive: true,
                    },
                ],
            })
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next()
    }

    private getLSPTextDocumentPositionParams(
        position: HoveredToken & RepoSpec & RevSpec & FileSpec & ResolvedRevSpec
    ): RepoSpec & RevSpec & ResolvedRevSpec & FileSpec & PositionSpec & ModeSpec {
        return {
            repoName: position.repoName,
            filePath: position.filePath,
            commitID: position.commitID,
            rev: position.rev,
            mode: getModeFromPath(FILE_PATH),
            position,
        }
    }

    public render(): JSX.Element {
        if (!this.state.blobOrError) {
            // Render placeholder for layout before content is fetched.
            return <div className="blob-page__placeholder">Loading...</div>
        }

        const hoverOverlayProps = this.adjustHoverOverlayPosition(this.target)

        return (
            <div className={this.props.className} ref={this.nextCodeIntellifyBlobElements}>
                <div className="code-header">
                    <span className="code-header__title">github.com/gorilla/mux/mux.go</span>
                    <span className="code-header__link">
                        <a href="https://sourcegraph.com/github.com/gorilla/mux/-/blob/mux.go">View full file</a>
                    </span>
                </div>
                {!isErrorLike(this.state.blobOrError) && (
                    <code
                        className={`blob__code blob__code--wrapped e2e-blob`}
                        ref={this.nextCodeViewElement}
                        dangerouslySetInnerHTML={{
                            __html: trimHTMLString(
                                this.state.blobOrError.highlight.html,
                                this.props.startLine - 1,
                                this.props.endLine + 1
                            ),
                        }}
                    />
                )}
                {this.state.hoverOverlayProps && (
                    <HoverOverlay
                        {...hoverOverlayProps}
                        hoverRef={this.nextOverlayElement}
                        extensionsController={this.props.extensionsController}
                        platformContext={this.props.platformContext}
                        location={this.props.location}
                        onCloseButtonClick={this.nextCloseButtonClick}
                        showCloseButton={false}
                        className={this.props.tooltipClass}
                    />
                )}
            </div>
        )
    }
    /**
     * This function adjusts the position of the hoverOverlay so that it does not overflow on the right side
     * of the viewport. If a hoverOverlay will exceed the viewport, this function will adjust the position
     * so that it aligns the right side of the hover overlay with the right side of the target element.
     *
     */
    private adjustHoverOverlayPosition(
        target: EventTarget | null
    ): HoverState<HoverContext, HoverMerged, ActionItemProps>['hoverOverlayProps'] {
        const viewPortEdge = window.innerWidth
        if (!this.state.hoverOverlayProps) {
            return undefined
        }
        if (!target) {
            return this.state.hoverOverlayProps
        }
        const { overlayPosition, ...rest } = this.state.hoverOverlayProps

        const targetBounds = (target as HTMLElement).getBoundingClientRect()
        let newOverlayPosition: { top: number; left: number } = overlayPosition!

        if (overlayPosition && viewPortEdge < targetBounds.left + 512 && targetBounds.right - 512 >= 0) {
            const containerWidth = (document.querySelector(
                this.props.parentElement
            ) as HTMLElement).parentElement!.getBoundingClientRect().width

            const parentWidth = (document.querySelector(
                this.props.parentElement
            ) as HTMLElement).getBoundingClientRect().width

            // One side of the total horizontal margin.
            const halfMarginWidth = (viewPortEdge - containerWidth) / 2
            // The difference between the viewport width and parent width. We need to subtract this because
            // `left` is relative to the parent, whereas targetBounds.right is relative to the viewport.
            const relativeElementDifference = viewPortEdge - parentWidth

            newOverlayPosition = {
                top: overlayPosition.top,
                // 512 is the width of a hoverOverlay.
                left: targetBounds.right - 512 - relativeElementDifference + halfMarginWidth,
            }
        }
        return { ...rest, overlayPosition: newOverlayPosition }
    }
}

/**
 * We can only fetch blobs as an entire file. For demo purposes, we only want to show part of the file.
 * This function trims the HTML string of the file that will be code-intellfied on the homepage to only show
 * the lines that we specify. It makes some assumptions for this specific use case, such as the presence
 * of a single table and tbody element in the html, so be careful when changing.
 */
function trimHTMLString(html: string, startLine: number, endLine: number): string {
    const domParser = new DOMParser()
    const doc = domParser.parseFromString(html, 'text/html')
    const startToRemove = doc.querySelectorAll(`tr:nth-child(n + 0):nth-child(-n + ${startLine})`)
    const endToRemove = doc.querySelectorAll(`tr:nth-child(n + ${endLine})`)

    const elementsToRemove = [...startToRemove, ...endToRemove]
    const tableEl = doc.querySelector('tbody')! // assume a single tbody element will exist in blob HTML

    for (const el of elementsToRemove) {
        tableEl.removeChild(el)
    }

    const xmlSerializer = new XMLSerializer()
    const tbl = doc.querySelector('table')! // assume a single table element will exist in blob HTML
    const trimmedHTMLString = xmlSerializer.serializeToString(tbl)

    return trimmedHTMLString
}
