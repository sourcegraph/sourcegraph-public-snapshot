import * as H from 'history'
import * as React from 'react'
import { Subscription } from 'rxjs'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { DiscussionsList } from '../../../discussions/DiscussionsList'
import { registerDiscussionsContributions } from './contributions'
import { DiscussionsCreate } from './DiscussionsCreate'
import { DiscussionsThread } from './DiscussionsThread'

interface Props extends ExtensionsControllerProps {
    repoID: GQL.ID
    repoName: string
    commitID: string
    rev: string | undefined
    filePath: string
    history: H.History
    location: H.Location
    compact: boolean
}

export class DiscussionsTree extends React.PureComponent<Props> {
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(registerDiscussionsContributions(this.props))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const hash = new URLSearchParams(this.props.location.hash.slice('#'.length))
        const threadID = hash.get('threadID') as GQL.ID
        const commentID = hash.get('commentID') as GQL.ID

        if (threadID && commentID) {
            return <DiscussionsThread threadID={threadID} commentID={commentID} {...this.props} />
        }
        if (threadID) {
            return <DiscussionsThread threadID={threadID} {...this.props} />
        }
        if (hash.get('createThread') === 'true') {
            return <DiscussionsCreate {...this.props} />
        }
        return <DiscussionsList {...this.props} />
    }
}
