import React, { useMemo, useState } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { QueryParameterProps } from '../../../../components/withQueryParameter/WithQueryParameter'
import { fetchDiscussionThreads } from '../../../../discussions/backend'

const LOADING: 'loading' = 'loading'

/**
 * React component props for children of {@link WithThreadsQueryResults}.
 */
export interface ThreadsQueryResultProps {
    /** The list of threads, loading, or an error. */
    threadsOrError: typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike
}

interface Props extends Pick<QueryParameterProps, 'query'> {
    children: (props: ThreadsQueryResultProps) => JSX.Element | null
}

/**
 * Wraps a component and provides a list of threads resulting from querying using the provided
 * `query` prop.
 */
export const WithThreadsQueryResults: React.FunctionComponent<Props> = ({ query, children }) => {
    const [threadsOrError, setThreadsOrError] = useState<typeof LOADING | GQL.IDiscussionThreadConnection | ErrorLike>(
        LOADING
    )

    // tslint:disable-next-line: no-floating-promises because this never throws
    useMemo(async () => {
        try {
            setThreadsOrError(await fetchDiscussionThreads({ query }).toPromise())
        } catch (err) {
            setThreadsOrError(asError(err))
        }
    }, [query])

    return children({ threadsOrError })
}
