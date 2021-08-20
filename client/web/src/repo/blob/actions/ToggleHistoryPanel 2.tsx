import * as H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { Tooltip } from '@sourcegraph/branded/src/components/tooltip/Tooltip'
import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    lprToRange,
    parseQueryAndHash,
    toPositionOrRangeQueryParameter,
    toViewStateHash,
} from '@sourcegraph/shared/src/util/url'

import { eventLogger } from '../../../tracking/eventLogger'
import { RepoHeaderContext } from '../../RepoHeader'
import { BlobPanelTabID } from '../panel/BlobPanel'

/**
 * A repository header action that toggles the visibility of the history panel.
 */
export class ToggleHistoryPanel extends React.PureComponent<
    {
        location: H.Location
        history: H.History
    } & RepoHeaderContext
> {
    private toggles = new Subject<boolean>()
    private subscriptions = new Subscription()

    /**
     * Reports the current visibility (derived from the location).
     */
    public static isVisible(location: H.Location): boolean {
        return parseQueryAndHash<BlobPanelTabID>(location.search, location.hash).viewState === 'history'
    }

    /**
     * Returns the location object (that can be passed to H.History's push/replace methods) that sets visibility to
     * the given value.
     */
    private static locationWithVisibility(location: H.Location, visible: boolean): H.LocationDescriptorObject {
        const parsedQuery = parseQueryAndHash<BlobPanelTabID>(location.search, location.hash)
        if (visible) {
            parsedQuery.viewState = 'history' // defaults to last-viewed tab, or first tab
        } else {
            delete parsedQuery.viewState
        }
        const lineRangeQueryParameter = toPositionOrRangeQueryParameter({ range: lprToRange(parsedQuery) })
        return {
            search: formatSearchParameters(
                addLineRangeQueryParameter(new URLSearchParams(location.search), lineRangeQueryParameter)
            ),
            hash: toViewStateHash(parsedQuery.viewState),
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.toggles.subscribe(() => {
                const visible = ToggleHistoryPanel.isVisible(this.props.location)
                eventLogger.log(visible ? 'HideHistoryPanel' : 'ShowHistoryPanel')
                this.props.history.push(ToggleHistoryPanel.locationWithVisibility(this.props.location, !visible))
                Tooltip.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+h' or 'opt+h'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.key === 'h'))
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
        const visible = ToggleHistoryPanel.isVisible(this.props.location)

        if (this.props.actionType === 'dropdown') {
            return (
                <ButtonLink className="btn repo-header__file-action" onSelect={this.onClick}>
                    <HistoryIcon className="icon-inline" />
                    <span>{visible ? 'Hide' : 'Show'} history (Alt+H/Opt+H)</span>
                </ButtonLink>
            )
        }
        return (
            <ButtonLink
                className="btn btn-icon repo-header__action"
                onSelect={this.onClick}
                data-tooltip={`${visible ? 'Hide' : 'Show'} history (Alt+H/Opt+H)`}
            >
                <HistoryIcon className="icon-inline" />
            </ButtonLink>
        )
    }

    private onClick = (): void => this.toggles.next()
}
