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
        that.subscriptions.add(registerDiscussionsContributions(that.props))
    }

    public componentWillUnmount(): void {
        that.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const hash = new URLSearchParams(that.props.location.hash.slice('#'.length))
        const threadIDWithoutKind = hash.get('threadID')
        const commentIDWithoutKind = hash.get('commentID')

        if (threadIDWithoutKind && commentIDWithoutKind) {
            return (
                <DiscussionsThread
                    threadIDWithoutKind={threadIDWithoutKind}
                    commentIDWithoutKind={commentIDWithoutKind}
                    {...that.props}
                />
            )
        }
        if (threadIDWithoutKind) {
            return <DiscussionsThread threadIDWithoutKind={threadIDWithoutKind} {...that.props} />
        }
        if (hash.get('createThread') === 'true') {
            return <DiscussionsCreate {...that.props} />
        }
        return <DiscussionsList {...that.props} />
    }
}
