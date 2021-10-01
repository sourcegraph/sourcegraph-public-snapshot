import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
} from '@sourcegraph/shared/src/graphql/graphql'

import { requestGraphQL } from '../../../backend/graphql'
import {
    DeleteLsifIndexResult,
    DeleteLsifIndexVariables,
    LsifIndexFields,
    LsifIndexResult,
    LsifIndexVariables,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment } from '../shared/backend'

export function fetchLsifIndex({ id }: { id: string }): Observable<LsifIndexFields | null> {
    const query = gql`
        query LsifIndex($id: ID!) {
            node(id: $id) {
                ...LsifIndexFields
            }
        }

        ${lsifIndexFieldsFragment}
    `

    return requestGraphQL<LsifIndexResult, LsifIndexVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node || node.__typename !== 'LSIFIndex') {
                throw new Error('No such LSIFIndex')
            }
            return node
        })
    )
}

export function deleteLsifIndex({ id }: { id: string }): Observable<void> {
    const query = gql`
        mutation DeleteLsifIndex($id: ID!) {
            deleteLSIFIndex(id: $id) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<DeleteLsifIndexResult, DeleteLsifIndexVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteLSIFIndex) {
                throw createInvalidGraphQLMutationResponseError('DeleteLsifIndex')
            }
        })
    )
}
