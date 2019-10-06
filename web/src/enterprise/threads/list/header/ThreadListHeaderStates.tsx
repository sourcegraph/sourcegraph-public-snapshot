import AlertOutlineIcon from 'mdi-react/AlertOutlineIcon'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import CheckIcon from 'mdi-react/CheckIcon'
import React from 'react'
import { isErrorLike, ErrorLike } from '../../../../../../shared/src/util/errors'
import { ListHeaderQueryLinksNav } from '../../../../components/listHeaderQueryLinks/ListHeaderQueryLinks'
import { QueryParameterProps } from '../../../../util/useQueryParameter'
import H from 'history'

const LOADING = 'loading' as const

interface Props extends Pick<QueryParameterProps, 'query'> {
    threads: typeof LOADING | GQL.IThreadConnection | GQL.IThreadOrThreadPreviewConnection | ErrorLike
    location: H.Location
}

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
                    count: threads.filters.openCount,
                    icon: AlertOutlineIcon,
                },
                {
                    label: 'closed',
                    queryField: 'is',
                    queryValues: ['closed'],
                    count: threads.filters.closedCount,
                    icon: CheckIcon,
                },
            ]}
            className="flex-1 nav"
        />
    ) : null
