import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'

interface Props {
    node: GQL.IDiscussionThread
}

export const ThreadsManageThreadsListItem: React.FunctionComponent<Props> = ({ node }) => (
    <li className="list-group-item">{node.title}</li>
)
