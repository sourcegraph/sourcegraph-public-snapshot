import * as React from 'react'

import * as H from 'history'
import LinkIcon from 'mdi-react/LinkIcon'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isInputElement } from '@sourcegraph/shared/src/util/dom'
import { Icon } from '@sourcegraph/wildcard'

import { replaceRevisionInURL } from '../../util/url'
import { RepoHeaderActionButtonLink } from '../components/RepoHeaderActions'
import { RepoHeaderContext } from '../RepoHeader'

/**
 * A repository header action that replaces the revision in the URL with the canonical 40-character
 * Git commit SHA.
 */
export class GoToPermalinkAction extends React.PureComponent<
    {
        /**
         * The current (possibly undefined or non-full-SHA) Git revision.
         */
        revision?: string

        /**
         * The commit SHA for the revision in the current location (URL).
         */
        commitID: string

        location: H.Location
        history: H.History
    } & RepoHeaderContext &
        TelemetryProps
> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Trigger the user presses 'y'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(
                    filter(
                        event =>
                            // 'y' shortcut (if no input element is focused)
                            event.key === 'y' && !!document.activeElement && !isInputElement(document.activeElement)
                    )
                )
                .subscribe(event => {
                    event.preventDefault()

                    // Replace the revision in the current URL with the new one and push to history.
                    this.props.history.push(this.permalinkURL)
                })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.props.revision === this.props.commitID) {
            return null // already at the permalink destination
        }

        if (this.props.actionType === 'dropdown') {
            return (
                <RepoHeaderActionButtonLink file={true} to={this.permalinkURL} onSelect={this.onClick.bind(this)}>
                    <Icon as={LinkIcon} />
                    <span>Permalink (with full Git commit SHA)</span>
                </RepoHeaderActionButtonLink>
            )
        }

        return (
            <RepoHeaderActionButtonLink
                className="btn-icon"
                file={false}
                to={this.permalinkURL}
                onSelect={this.onClick.bind(this)}
                data-tooltip="Permalink (with full Git commit SHA)"
            >
                <Icon as={LinkIcon} />
            </RepoHeaderActionButtonLink>
        )
    }

    private onClick(): void {
        this.props.telemetryService.log('PermalinkClicked', {
            repoName: this.props.repoName,
            commitID: this.props.commitID,
        })
    }

    private get permalinkURL(): string {
        return replaceRevisionInURL(
            this.props.location.pathname + this.props.location.search + this.props.location.hash,
            this.props.commitID
        )
    }
}
