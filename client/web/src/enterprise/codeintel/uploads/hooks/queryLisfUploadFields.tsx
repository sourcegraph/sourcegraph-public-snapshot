import { ApolloClient } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { ErrorLike } from '@sourcegraph/common'
import { fromObservableQuery, gql, getDocumentNode } from '@sourcegraph/http-client'

import { LsifUploadFields, LsifUploadResult } from '../../../../graphql-operations'

import { lsifUploadFieldsFragment } from './types'

const LSIF_UPLOAD_FIELDS = gql`
    query LsifUpload($id: ID!) {
        node(id: $id) {
            ...LsifUploadFields
        }
    }

    ${lsifUploadFieldsFragment}
`

const LSIF_UPLOAD_POLL_INTERVAL = 5000

export const queryLisfUploadFields = (
    id: string,
    client: ApolloClient<object>
): Observable<LsifUploadFields | ErrorLike | null | undefined> =>
    fromObservableQuery(
        client.watchQuery<LsifUploadResult>({
            query: getDocumentNode(LSIF_UPLOAD_FIELDS),
            variables: { id },
            pollInterval: LSIF_UPLOAD_POLL_INTERVAL,
        })
    ).pipe(
        map(({ data }) => data),
        map(({ node }) => {
            if (!node || node.__typename !== 'LSIFUpload') {
                throw new Error('No such LSIFUpload')
            }
            return node
        })
    )
