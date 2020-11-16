import {
    createHoverifier,
    DOMFunctions,
    findPositionsFromEvents,
    HoveredToken,
    HoverState,
    Hoverifier as HoverifierInstance,
} from '@sourcegraph/codeintellify'
import React from 'react'
import { fromEvent, Observable, Subject, Subscription } from 'rxjs'
import { filter, map, switchMap, tap, withLatestFrom } from 'rxjs/operators'
import { ActionItemAction } from '../../../shared/src/actions/ActionItem'
import { HoverMerged } from '../../../shared/src/api/client/types/hover'
import { Controller } from '../../../shared/src/extensions/controller'
import { getHoverActions } from '../../../shared/src/hover/actions'
import { HoverContext, HoverOverlayProps } from '../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../shared/src/languages'
import { PlatformContext } from '../../../shared/src/platform/context'
import { isDefined, property } from '../../../shared/src/util/types'
import { AbsoluteRepoFile, LineOrPositionOrRange, ModeSpec, UIPositionSpec } from '../../../shared/src/util/url'
import { getDocumentHighlights, getHover } from '../backend/features'
import { HoverThresholdProps } from '../repo/RepoContainer'
import { WebHoverOverlay } from './shared'

interface HoverifierProps<P> extends Omit<HoverOverlayProps, 'showCloseButton'>, HoverThresholdProps {
    /** Function component as child, called with  */
    children: React.FunctionComponent<HoverifiedProps<P>>

    /**
     * Value to pass through from `Hoverifier`s parent to its child.
     * Useful when you want to memoize the function, but need to access changing values.
     */
    passthroughProps: P

    /** A file at an exact commit */
    absoluteRepoFile: AbsoluteRepoFile

    /** Whether or not hover tooltips can be pinned */
    pinningEnabled: boolean

    /**
     * A collection of methods needed to tell codeintellify how to look at the DOM. These are required for
     * ensuring that we don't rely on any sort of specific DOM structure.
     */
    domFunctions: DOMFunctions

    /** Should emit parsed positions found in the URL */
    locationPositions: Observable<LineOrPositionOrRange>

    /** Called when line selections have changed */
    onLineSelection?: (event: MouseEvent, hoverState: HoverState<HoverContext, HoverMerged, ActionItemAction>) => void

    /** Called when hover state updates */
    onHoverStateUpdate?: (hoverState: HoverState<HoverContext, HoverMerged, ActionItemAction>) => void

    /** Platform-specific data and methods shared by multiple Sourcegraph components. */
    platformContext: PlatformContext

    /** The client, which is used to communicate with and manage extensions. */
    extensionsController: Controller
}

/**
 * Props passed to `Hoverifier`s child
 */
export interface HoverifiedProps<P> {
    /**
     * The `hoverifier` instance created by `Hoverifier`
     */
    hoverifier: HoverifierInstance<HoverContext, HoverMerged, ActionItemAction>

    /**
     * Optional value to pass through from `Hoverifier`s parent to its child.
     * Useful when you want to memoize the function, but need to access changing values.
     */
    passthroughProps: P

    /**
     * The hover overlay managed by `Hoverifier`. You must render this overlay in the function child.
     */
    overlay: React.ReactNode

    /** Should emit whenever the ref callback for the blob element is called */
    nextBlobElement: (blobElement: HTMLElement | null) => void

    /** Should emit whenever the ref callback for the code element is called */
    nextCodeViewElement: (codeView: HTMLElement | null) => void
}

interface State extends HoverState<HoverContext, HoverMerged, ActionItemAction> {
    hoverifier: HoverifierInstance<HoverContext, HoverMerged, ActionItemAction>
}

/**
 * Creates a `hoverifier` instance and "hoverifies" a code view.
 *
 * @template P Type of value to to pass from `Hoverifier`s parent to its child
 */
export class Hoverifier<P = undefined> extends React.Component<HoverifierProps<P>, State> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<HoverifierProps<P>>()

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null): void => this.hoverOverlayElements.next(element)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent): void => this.closeButtonClicks.next(event)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (blobElement: HTMLElement | null): void => this.blobElements.next(blobElement)

    /** Emits whenever the ref callback for the code element is called */
    private codeViewElements = new Subject<HTMLElement | null>()
    private nextCodeViewElement = (codeView: HTMLElement | null): void => this.codeViewElements.next(codeView)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: HoverifierProps<P>) {
        super(props)

        const { locationPositions } = this.props

        // Create hoverifier instance
        const hoverifier = createHoverifier<HoverContext, HoverMerged, ActionItemAction>({
            closeButtonClicks: this.closeButtonClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.blobElements),
                map(([, hoverOverlayElement, blobElement]) => ({
                    hoverOverlayElement,
                    relativeElement: blobElement,
                })),
                filter(property('relativeElement', isDefined)),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(property('hoverOverlayElement', isDefined))
            ),
            getHover: position => getHover(getLSPTextDocumentPositionParameters(position), this.props),
            getDocumentHighlights: position =>
                getDocumentHighlights(getLSPTextDocumentPositionParameters(position), this.props),
            getActions: context => getHoverActions(this.props, context),
            pinningEnabled: this.props.pinningEnabled,
        })
        this.subscriptions.add(hoverifier)

        this.state = { hoverifier }

        // Hoverify code view
        this.subscriptions.add(
            hoverifier.hoverify({
                positionEvents: this.codeViewElements.pipe(
                    filter(isDefined),
                    findPositionsFromEvents({ domFunctions: this.props.domFunctions })
                ),
                positionJumps: locationPositions.pipe(
                    withLatestFrom(this.codeViewElements, this.blobElements),
                    map(([position, codeView, scrollElement]) => ({
                        position,
                        // locationPositions is derived from componentUpdates,
                        // so these elements are guaranteed to have been rendered.
                        codeView: codeView!,
                        scrollElement: scrollElement!,
                    }))
                ),
                resolveContext: () => this.props.absoluteRepoFile,
                dom: this.props.domFunctions,
            })
        )

        // Call `onLineSelection` callback on click events
        this.subscriptions.add(
            this.codeViewElements
                .pipe(
                    filter(isDefined),
                    switchMap(codeView => fromEvent<MouseEvent>(codeView, 'click')),
                    // Ignore click events caused by the user selecting text
                    filter(() => !window.getSelection()?.toString())
                )
                .subscribe(event => {
                    // Prevent selecting text on shift click (click+drag to select will still work)
                    // Note that this is only called if the selection was empty initially (see above),
                    // so this only clears a selection caused by this click.
                    window.getSelection()!.removeAllRanges()
                    this.props.onLineSelection?.(event, hoverifier.hoverState)
                })
        )

        this.subscriptions.add(
            hoverifier.hoverStateUpdates.subscribe(update => {
                this.props.onHoverStateUpdate?.(update)
                this.setState(update)
            })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        // Don't call React.createElement in order to force React to diff even when function child has a different identity.
        // Otherwise, React may unnecessarily recreate DOM nodes, causing undefined behavior in `hoverifier`
        // https://www.huy.dev/2017-01-avoid-unnecessary-remounting-react/
        return this.props.children({
            passthroughProps: this.props.passthroughProps,
            hoverifier: this.state.hoverifier,
            nextBlobElement: this.nextBlobElement,
            nextCodeViewElement: this.nextCodeViewElement,
            overlay: this.state.hoverOverlayProps && (
                <WebHoverOverlay
                    {...this.props}
                    {...this.state.hoverOverlayProps}
                    hoverRef={this.nextOverlayElement}
                    onCloseButtonClick={this.nextCloseButtonClick}
                />
            ),
        })
    }
}

export function getLSPTextDocumentPositionParameters(
    hoveredToken: HoveredToken & AbsoluteRepoFile
): AbsoluteRepoFile & UIPositionSpec & ModeSpec {
    return {
        repoName: hoveredToken.repoName,
        revision: hoveredToken.revision,
        filePath: hoveredToken.filePath,
        commitID: hoveredToken.commitID,
        position: hoveredToken,
        mode: getModeFromPath(hoveredToken.filePath || ''),
    }
}
