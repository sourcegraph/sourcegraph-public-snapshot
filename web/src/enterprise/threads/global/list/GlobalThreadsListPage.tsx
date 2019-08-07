import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { ThreadsList } from '../../list/ThreadsList'
import { useThreads } from '../../list/useThreads'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all threads.
 */
export const GlobalThreadsListPage: React.FunctionComponent<Props> = props => {
    const threads = useThreads()
    return <ThreadsList {...props} threads={threads} />
}
