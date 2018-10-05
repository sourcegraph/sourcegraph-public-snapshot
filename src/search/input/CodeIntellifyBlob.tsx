import {
    createHoverifier,
    findPositionsFromEvents,
    HoveredToken,
    HoveredTokenContext,
    HoverOverlay,
    HoverState,
} from '@sourcegraph/codeintellify'
import { getTokenAtPosition } from '@sourcegraph/codeintellify/lib/token_position'
import * as H from 'history'
import * as React from 'react'
import { Link, LinkProps } from 'react-router-dom'
import { Subject, Subscription } from 'rxjs'
import { catchError, filter, map, withLatestFrom } from 'rxjs/operators'
import { Position } from 'vscode-languageserver-types'
import { getHover, getJumpURL } from '../../backend/features'
import * as GQL from '../../backend/graphqlschema'
import { LSPTextDocumentPositionParams } from '../../backend/lsp'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import { ExtensionsControllerProps } from '../../extensions/ExtensionsClientCommonContext'
import { fetchBlob } from '../../repo/blob/BlobPage'
import { getModeFromPath } from '../../util'
import { ErrorLike, isErrorLike } from '../../util/errors'
import { isDefined, propertyIsDefined } from '../../util/types'

interface Props extends ExtensionsControllerProps, ExtensionsDocumentsProps {
    history: H.History
    className: string
    startLine: number
    endLine: number
    parentElement: string

    overlayPortal?: HTMLElement
    tooltipClass: string
    defaultHoverPosition: Position
}

interface State extends HoverState {
    /**
     * The blob data or error that happened.
     * undefined while loading.
     */
    blobOrError?: GQL.IGitBlob | ErrorLike
    target?: EventTarget
}

const LinkComponent = (props: LinkProps) => <Link {...props} />

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the blame annotation (a.blame or span.blame__contents), return null.
        if (
            target.classList.contains('blame') ||
            target.classList.contains('blame__contents') ||
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

const REPO_PATH = 'github.com/gorilla/mux'
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

    /** Emits when the go to definition button was clicked */
    private goToDefinitionClicks = new Subject<MouseEvent>()
    private nextGoToDefinitionClick = (event: MouseEvent) => this.goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent) => this.closeButtonClicks.next(event)

    private subscriptions = new Subscription()

    private componentUpdates = new Subject<void>()

    private target: EventTarget | null = null

    constructor(props: Props) {
        super(props)
        this.state = {}

        const hoverifier = createHoverifier({
            closeButtonClicks: this.closeButtonClicks,
            goToDefinitionClicks: this.goToDefinitionClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            pushHistory: path => this.props.history.push(path),
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
            fetchHover: hoveredToken => getHover(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
            fetchJumpURL: hoveredToken => getJumpURL(this.getLSPTextDocumentPositionParams(hoveredToken), this.props),
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
                    repoPath: REPO_PATH,
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
            repoPath: REPO_PATH,
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
            this.props.extensionsOnVisibleTextDocumentsChange([
                {
                    uri: `git://github.com/gorilla/mux?9e1f5955c0d22b55d9e20d6faa28589f83b2faca#mux.go`,
                    languageId: 'go',
                    text: '',
                },
            ])
        )
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next()
    }

    private getLSPTextDocumentPositionParams(
        position: HoveredToken & HoveredTokenContext
    ): LSPTextDocumentPositionParams {
        return {
            repoPath: position.repoPath,
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
                        linkComponent={LinkComponent}
                        hoverRef={this.nextOverlayElement}
                        onGoToDefinitionClick={this.nextGoToDefinitionClick}
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
    private adjustHoverOverlayPosition(target: EventTarget | null): HoverState['hoverOverlayProps'] {
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
