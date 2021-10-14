import { ApolloError, MutationFunctionOptions, FetchResult, ApolloClient, useMutation } from '@apollo/client'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { fromObservableQuery } from '@sourcegraph/shared/src/graphql/apollo'
import { gql, getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import { ErrorLike } from '@sourcegraph/shared/src/util/errors'

import {
    DeleteLsifIndexResult,
    DeleteLsifIndexVariables,
    LsifIndexFields,
    LsifIndexResult,
    Exact,
} from '../../../graphql-operations'
import { lsifIndexFieldsFragment } from '../shared/backend'

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

type DeleteLsifIndexResults = Promise<FetchResult<DeleteLsifIndexResult, Record<string, any>, Record<string, any>>>

interface UseDeleteLsifIndexResult {
    handleDeleteLsifIndex: (
        options?:
            | MutationFunctionOptions<
                  DeleteLsifIndexResult,
                  Exact<{
                      id: string
                  }>
              >
            | undefined
    ) => DeleteLsifIndexResults
    deleteError: ApolloError | undefined
}

const DELETE_LSIF_INDEX = gql`
    mutation DeleteLsifIndex($id: ID!) {
        deleteLSIFIndex(id: $id) {
            alwaysNil
        }
    }
`

export const useDeleteLsifIndex = (): UseDeleteLsifIndexResult => {
    const [handleDeleteLsifIndex, { error }] = useMutation<DeleteLsifIndexResult, DeleteLsifIndexVariables>(
        getDocumentNode(DELETE_LSIF_INDEX)
    )

    return {
        handleDeleteLsifIndex,
        deleteError: error,
    }
}
