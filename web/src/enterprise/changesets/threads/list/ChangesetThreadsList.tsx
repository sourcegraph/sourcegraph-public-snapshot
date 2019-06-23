import H from 'history'
import React from 'react'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadsAreaContext } from '../../../threads/global/ThreadsArea'
import { ThreadsList } from '../../../threads/list/ThreadsList'
import { ChangesetThreadsListHeader } from './ChangesetThreadsListHeader'

interface Props extends QueryParameterProps, Pick<ThreadsAreaContext, 'type' | 'authenticatedUser'> {
    history: H.History
    location: H.Location
}

/**
 * The list of changeset threads.
 */
export const ChangesetThreadsList: React.FunctionComponent<Props> = props => (
    <>
        <ChangesetThreadsListHeader {...props} />
        <ThreadsList {...props} itemCheckboxes={false} />
    </>
)
