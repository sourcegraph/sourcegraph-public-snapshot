import * as H from 'history'
import * as React from 'react'
import * as GQL from '../../../backend/graphqlschema'
import { DiscussionsList } from '../../../discussions/DiscussionsList'
import { DiscussionsCreate } from './DiscussionsCreate'
import { DiscussionsThread } from './DiscussionsThread'

interface Props {
    repoID: GQL.ID
    repoPath: string
    commitID: string
    rev: string | undefined
    filePath: string
    history: H.History
    location: H.Location
    user: GQL.IUser | null
}

export class DiscussionsTree extends React.PureComponent<Props> {
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
