import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
} from '../../../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../../../backend/graphql'
import {
    DeleteLsifIndexResult,
    DeleteLsifIndexVariables,
    DeleteLsifUploadResult,
    DeleteLsifUploadVariables,
    LsifIndexFields,
    LsifIndexResult,
    LsifIndexVariables,
    LsifUploadFields,
    LsifUploadResult,
    LsifUploadVariables,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment, lsifUploadFieldsFragment } from '../shared/backend'

export function fetchLsifUpload({ id }: { id: string }): Observable<LsifUploadFields | null> {
    const query = gql`
        query LsifUpload($id: ID!) {
            node(id: $id) {
                ...LsifUploadFields
            }
        }

        ${lsifUploadFieldsFragment}
    `

    return requestGraphQL<LsifUploadResult, LsifUploadVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(({ node }) => {
            if (!node) {
                throw new Error('No such LSIFUpload')
            }
            return node
        })
    )
}

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
            if (!node) {
                throw new Error('No such LSIFIndex')
            }
            return node
        })
    )
}

export function deleteLsifUpload({ id }: { id: string }): Observable<void> {
    const query = gql`
        mutation DeleteLsifUpload($id: ID!) {
            deleteLSIFUpload(id: $id) {
                alwaysNil
            }
        }
    `

    return requestGraphQL<DeleteLsifUploadResult, DeleteLsifUploadVariables>(query, { id }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteLSIFUpload) {
                throw createInvalidGraphQLMutationResponseError('DeleteLsifUpload')
            }
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
