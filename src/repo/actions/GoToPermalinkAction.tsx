import * as H from 'history'
import LinkIcon from 'mdi-react/LinkIcon'
import * as React from 'react'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { replaceRevisionInURL } from '..'
import { ActionItem } from '../../components/ActionItem'

/**
 * A repository header action that replaces the revision in the URL with the canonical 40-character
 * Git commit SHA.
 */
export class GoToPermalinkAction extends React.PureComponent<{
    /**
     * The current (possibly undefined or non-full-SHA) Git revision.
     */
    rev?: string

    /**
     * The commit SHA for the revision in the current location (URL).
     */
    commitID: string

    location: H.Location
    history: H.History
}> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        // Trigger the user presses 'y'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(
                    filter(
                        event =>
                            // 'y' shortcut (if no input element is focused)
                            event.key === 'y' && !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)
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
        if (this.props.rev === this.props.commitID) {
            return null // already at the permalink destination
        }

        return (
            <ActionItem to={this.permalinkURL} data-tooltip="Permalink (with full Git commit SHA)">
                <LinkIcon className="icon-inline" />
            </ActionItem>
        )
    }

    private get permalinkURL(): string {
        return replaceRevisionInURL(
            this.props.location.pathname + this.props.location.search + this.props.location.hash,
            this.props.commitID
        )
    }
}
