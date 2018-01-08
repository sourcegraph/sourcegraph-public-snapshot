import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs/Subscription'
import { currentUser } from '../../auth'
import { hasTagRecursive } from '../../settings/tags'
import { eventLogger } from '../../tracking/eventLogger'
import { parseHash, toEditorURL } from '../../util/url'

/**
 * A repository header action that opens the current file in Sourcegraph Editor.
 */
export class OpenInEditorAction extends React.PureComponent<
    {
        location: H.Location
        repoPath: string
        commitID: string
        filePath: string
        customEditorURL?: string
    },
    { editorBeta: boolean }
> {
    public state = { editorBeta: false }

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            currentUser.subscribe(user => {
                this.setState({ editorBeta: hasTagRecursive(user, 'editor-beta') })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.editorBeta) {
            return null
        }

        return (
            <a
                href={
                    this.props.customEditorURL ||
                    toEditorURL(
                        this.props.repoPath,
                        this.props.commitID,
                        this.props.filePath,
                        parseHash(this.props.location.hash)
                    )
                }
                className="composite-container__header-action"
                title="Open in Sourcegraph Editor"
                onClick={this.onClick}
            >
                <PencilIcon className="icon-inline" />
                <span className="composite-container__header-action-text">Edit</span>
            </a>
        )
    }

    private onClick: React.MouseEventHandler<HTMLAnchorElement> = () => {
        eventLogger.log('OpenInNativeAppClicked')
    }
}
