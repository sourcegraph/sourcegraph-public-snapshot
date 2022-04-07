import * as React from 'react'

import * as H from 'history'
import HistoryIcon from 'mdi-react/HistoryIcon'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import {
    addLineRangeQueryParameter,
    formatSearchParameters,
    lprToRange,
    toPositionOrRangeQueryParameter,
    toViewStateHash,
} from '@sourcegraph/common'
import { parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { TooltipController, Icon } from '@sourcegraph/wildcard'

import { eventLogger } from '../../../tracking/eventLogger'
import { RepoHeaderActionButtonLink } from '../../components/RepoHeaderActions'
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
                TooltipController.forceUpdate()
            })
        )

        // Toggle when the user presses 'alt+h' or 'opt+h'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.altKey && event.code === 'KeyH'))
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
                <RepoHeaderActionButtonLink file={true} onSelect={this.onClick}>
                    <Icon as={HistoryIcon} />
                    <span>{visible ? 'Hide' : 'Show'} history (Alt+H/Opt+H)</span>
                </RepoHeaderActionButtonLink>
            )
        }
        return (
            <RepoHeaderActionButtonLink
                className="btn-icon"
                file={false}
                onSelect={this.onClick}
                data-tooltip={`${visible ? 'Hide' : 'Show'} history (Alt+H/Opt+H)`}
            >
                <Icon as={HistoryIcon} />
            </RepoHeaderActionButtonLink>
        )
    }

    private onClick = (): void => this.toggles.next()
}
