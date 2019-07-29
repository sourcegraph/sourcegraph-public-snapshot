import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { ThreadsList } from '../../list/ThreadsList'
import { useThreads } from '../../list/useThreads'
import { GlobalNewThreadDropdownButton } from './GlobalNewThreadDropdownButton'

interface Props extends ExtensionsControllerNotificationProps {}

/**
 * A list of all threads.
 */
export const GlobalThreadsListPage: React.FunctionComponent<Props> = props => {
    const threads = useThreads()
    return (
        <>
            <GlobalNewThreadDropdownButton className="mb-3" />
            <ThreadsList {...props} threads={threads} />
        </>
    )
}
