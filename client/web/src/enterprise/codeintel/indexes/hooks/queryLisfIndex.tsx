import { ApolloClient } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { ErrorLike } from '@sourcegraph/common'
import { fromObservableQuery, gql, getDocumentNode } from '@sourcegraph/http-client'

import { LsifIndexFields, LsifIndexResult } from '../../../../graphql-operations'

import { lsifIndexFieldsFragment } from './types'

const LSIF_INDEX_FIELDS = gql`
    query LsifIndex($id: ID!) {
        node(id: $id) {
            ...LsifIndexFields
        }
    }

    ${lsifIndexFieldsFragment}
`

const LSIF_INDEX_POLL_INTERVAL = 5000

export const queryLisfIndex = (
    id: string,
    client: ApolloClient<object>
): Observable<LsifIndexFields | ErrorLike | null | undefined> =>
    fromObservableQuery(
        client.watchQuery<LsifIndexResult>({
            query: getDocumentNode(LSIF_INDEX_FIELDS),
            variables: { id },
            pollInterval: LSIF_INDEX_POLL_INTERVAL,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node || node.__typename !== 'LSIFIndex') {
                throw new Error('No such LSIFIndex')
            }
            return node
        })
    )
