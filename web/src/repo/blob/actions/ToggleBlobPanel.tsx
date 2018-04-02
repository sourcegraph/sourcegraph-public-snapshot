import NewsIcon from '@sourcegraph/icons/lib/News'
import * as H from 'history'
import * as React from 'react'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { filter } from 'rxjs/operators/filter'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { eventLogger } from '../../../tracking/eventLogger'
import { lprToRange, parseHash, toPositionOrRangeHash, toViewStateHashComponent } from '../../../util/url'

/**
 * A repository header action that toggles the visibility of the blob panel.
 */
export class ToggleBlobPanel extends React.PureComponent<{
    location: H.Location
    history: H.History
}> {
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    /**
     * Reports the current visibility (derived from the location).
     */
    public static isVisible(location: H.Location): boolean {
        return !!parseHash(location.hash).viewState
    }

    /**
     * Returns the location object (that can be passed to H.History's push/replace methods) that sets visibility to
     * the given value.
     */
    private static locationWithVisibility(location: H.Location, visible: boolean): H.LocationDescriptorObject {
        const hash = parseHash(location.hash)
        if (visible) {
            hash.viewState = 'references'
        } else {
            delete hash.viewState
        }
        return { hash: toPositionOrRangeHash({ range: lprToRange(hash) }) + toViewStateHashComponent(hash.viewState) }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.toggles.subscribe(() => {
                const visible = ToggleBlobPanel.isVisible(this.props.location)
                eventLogger.log(visible ? 'HidePanel' : 'ShowPanel')
                this.props.history.push(ToggleBlobPanel.locationWithVisibility(this.props.location, !visible))
                Tooltip.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+x'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                // Opt/alt+x shortcut
                .pipe(filter(event => event.altKey && event.keyCode === 88))
                .subscribe(event => {
                    event.preventDefault()
                    this.toggles.next()
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const visible = ToggleBlobPanel.isVisible(this.props.location)
        return (
            <button
                className="btn btn-link btn-sm composite-container__header-action"
                onClick={this.onClick}
                data-tooltip={`${visible ? 'Hide' : 'Show'} panel (Alt+X/Opt+X)`}
            >
                <NewsIcon className="icon-inline" />
            </button>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = () => this.toggles.next()
}
