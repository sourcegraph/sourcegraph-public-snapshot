import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ListHeaderQueryLinksNav } from '../../threadsOLD/components/ListHeaderQueryLinks'
import { ThreadListHeaderContext } from './ThreadListHeader'

const LOADING = 'loading' as const

interface Props extends ThreadListHeaderContext {}

/**
 * A list of thread states with counts for the thread list header.
 */
export const ThreadListHeaderStates: React.FunctionComponent<Props> = ({ threads, ...props }) =>
    threads !== LOADING && !isErrorLike(threads) ? (
        <ListHeaderQueryLinksNav
            {...props}
            links={[
                {
                    label: 'open',
                    queryField: 'is',
                    queryValues: ['open'],
                    count: threads.totalCount,
                    icon: AlertOutlineIcon,
                },
                {
                    label: 'closed',
                    queryField: 'is',
                    queryValues: ['closed'],
                    count: 0,
                    icon: CheckIcon,
                },
            ]}
            className="flex-1 nav"
        />
    ) : null
