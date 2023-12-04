import type { ApolloClient } from '@apollo/client'
import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import type { ErrorLike } from '@sourcegraph/common'
import { fromObservableQuery, getDocumentNode, gql } from '@sourcegraph/http-client'

import type { PreciseIndexFields, PreciseIndexResult } from '../../../../graphql-operations'

import { preciseIndexFieldsFragment } from './types'

const PRECISE_INDEX_FIELDS = gql`
    query PreciseIndex($id: ID!) {
        node(id: $id) {
            ...PreciseIndexFields
        }
    }

    ${preciseIndexFieldsFragment}
`

const POLL_INTERVAL = 5000

export const queryPreciseIndex = (
    id: string,
    client: ApolloClient<object>
): Observable<PreciseIndexFields | ErrorLike | null | undefined> =>
    fromObservableQuery(
        client.watchQuery<PreciseIndexResult>({
            query: getDocumentNode(PRECISE_INDEX_FIELDS),
            variables: { id },
            pollInterval: POLL_INTERVAL,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node || node.__typename !== 'PreciseIndex') {
                throw new Error('No such precise index')
            }
            return node
        })
    )
