import * as H from 'history'
import LinkIcon from 'mdi-react/LinkIcon'
import * as React from 'react'
import { Subscription } from 'rxjs'

import { BuiltInCommand } from '@sourcegraph/shared/src/commandPalette/v2/store'

import { ButtonLink } from '@sourcegraph/shared/src/components/LinkOrButton'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { replaceRevisionInURL } from '../../util/url'
import { RepoHeaderContext } from '../RepoHeader'
import { CommandItem } from '@sourcegraph/shared/src/commandPalette/v2/components/CommandResult'

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

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (this.props.revision === this.props.commitID) {
            return null // already at the permalink destination
        }

        if (this.props.actionType === 'dropdown') {
            return (
                <>
                    <BuiltInCommand commandItem={this.commandItem} />
                    <ButtonLink
                        className="btn repo-header__file-action"
                        to={this.permalinkURL}
                        onSelect={this.onClick.bind(this)}
                    >
                        <LinkIcon className="icon-inline" />
                        <span>Permalink (with full Git commit SHA)</span>
                    </ButtonLink>
                </>
            )
        }

        return (
            <>
                <BuiltInCommand commandItem={this.commandItem} />
                <ButtonLink
                    to={this.permalinkURL}
                    data-tooltip="Permalink (with full Git commit SHA)"
                    onSelect={this.onClick.bind(this)}
                    className="btn btn-icon repo-header__action"
                >
                    <LinkIcon className="icon-inline" />
                </ButtonLink>
            </>
        )
    }

    private onClick(): void {
        this.props.telemetryService.log('PermalinkClicked', {
            repoName: this.props.repoName,
            commitID: this.props.commitID,
        })
    }

    private commandItem: CommandItem = {
        id: 'expandURL',
        title: 'Expand URL to its canonical form (on file or tree page)',
        keybindings: [{ ordered: ['y'] }],
        onClick: () => {
            this.props.history.push(this.permalinkURL)
        },
    }

    private get permalinkURL(): string {
        return replaceRevisionInURL(
            this.props.location.pathname + this.props.location.search + this.props.location.hash,
            this.props.commitID
        )
    }
}
