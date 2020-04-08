import * as H from 'history'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { ChatIcon } from '../../../../../shared/src/components/icons'
import { LinkOrButton } from '../../../../../shared/src/components/LinkOrButton'
import {
    lprToRange,
    parseHash,
    toPositionOrRangeHash,
    toViewStateHashComponent,
} from '../../../../../shared/src/util/url'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { eventLogger } from '../../../tracking/eventLogger'
import { BlobPanelTabID } from '../panel/BlobPanel'
/**
 * A repository header action that toggles the visibility of the discussions panel.
 */
export class ToggleDiscussionsPanel extends React.PureComponent<{
    location: H.Location
    history: H.History
}> {
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    /**
     * Reports the current visibility (derived from the location).
     */
    public static isVisible(location: H.Location): boolean {
        return parseHash<BlobPanelTabID>(location.hash).viewState === 'discussions'
    }

    /**
     * Returns the location object (that can be passed to H.History's push/replace methods) that sets visibility to
     * the given value.
     */
    private static locationWithVisibility(location: H.Location, visible: boolean): H.LocationDescriptorObject {
        const hash = parseHash<BlobPanelTabID>(location.hash)
        if (visible) {
            hash.viewState = 'discussions' // defaults to last-viewed tab, or first tab
        } else {
            delete hash.viewState
        }
        return { hash: toPositionOrRangeHash({ range: lprToRange(hash) }) + toViewStateHashComponent(hash.viewState) }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.toggles.subscribe(() => {
                const visible = ToggleDiscussionsPanel.isVisible(this.props.location)
                eventLogger.log(visible ? 'HideDiscussionsPanel' : 'ShowDiscussionsPanel')
                this.props.history.push(ToggleDiscussionsPanel.locationWithVisibility(this.props.location, !visible))
                Tooltip.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+d' or 'opt+d'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.key === 'd'))
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
        const visible = ToggleDiscussionsPanel.isVisible(this.props.location)
        return (
            <LinkOrButton
                onSelect={this.onClick}
                data-tooltip={`${visible ? 'Hide' : 'Show'} discussions (Alt+D/Opt+D)`}
            >
                <ChatIcon className="icon-inline" />
            </LinkOrButton>
        )
    }

    private onClick = (): void => this.toggles.next()
}
