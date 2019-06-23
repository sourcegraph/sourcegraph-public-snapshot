import H from 'history'
import React from 'react'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThreadsAreaContext } from '../../../threads/global/ThreadsArea'
import { ThreadsList } from '../../../threads/list/ThreadsList'
import { CheckThreadsListHeader } from './CheckThreadsListHeader'

interface Props extends QueryParameterProps, Pick<ThreadsAreaContext, 'type' | 'authenticatedUser'> {
    history: H.History
    location: H.Location
}

/**
 * The list of check threads.
 */
export const CheckThreadsList: React.FunctionComponent<Props> = props => (
    <>
        <CheckThreadsListHeader {...props} />
        <ThreadsList {...props} itemCheckboxes={false} />
    </>
)
