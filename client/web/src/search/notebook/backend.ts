import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import { requestGraphQL } from '../../backend/graphql'
import {
    CreateNotebookResult,
    CreateNotebookVariables,
    DeleteNotebookResult,
    DeleteNotebookVariables,
    FetchNotebookResult,
    FetchNotebookVariables,
    ListNotebooksResult,
    ListNotebooksVariables,
    Maybe,
    NotebookFields,
    Scalars,
    UpdateNotebookResult,
    UpdateNotebookVariables,
} from '../../graphql-operations'

const notebooksFragment = gql`
    fragment NotebookFields on Notebook {
        __typename
        id
        title
        creator {
            username
        }
        createdAt
        updatedAt
        public
        viewerCanManage
        blocks {
            ... on MarkdownBlock {
                __typename
                id
                markdownInput
            }
            ... on QueryBlock {
                __typename
                id
                queryInput
            }
            ... on FileBlock {
                __typename
                id
                fileInput {
                    __typename
                    repositoryName
                    filePath
                    revision
                    lineRange {
                        __typename
                        startLine
                        endLine
                    }
                }
            }
        }
    }
`

const fetchNotebooksQuery = gql`
    query ListNotebooks(
        $first: Int!
        $after: String
        $orderBy: NotebooksOrderBy
        $descending: Boolean
        $creatorUserID: ID
        $query: String
    ) {
        notebooks(
            first: $first
            after: $after
            orderBy: $orderBy
            descending: $descending
            creatorUserID: $creatorUserID
            query: $query
        ) {
            nodes {
                ...NotebookFields
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }
    ${notebooksFragment}
`

export function fetchNotebooks({
    first,
    creatorUserID,
    query,
    after,
    orderBy,
    descending,
}: {
    first: number
    query?: string
    creatorUserID?: Maybe<Scalars['ID']>
    after?: string
    orderBy?: GQL.NotebooksOrderBy
    descending?: boolean
}): Observable<ListNotebooksResult['notebooks']> {
    return requestGraphQL<ListNotebooksResult, ListNotebooksVariables>(fetchNotebooksQuery, {
        first,
        after: after ?? null,
        query: query ?? null,
        creatorUserID: creatorUserID ?? null,
        orderBy: orderBy ?? GQL.NotebooksOrderBy.NOTEBOOK_UPDATED_AT,
        descending: descending ?? false,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.notebooks)
    )
}

const fetchNotebookQuery = gql`
    query FetchNotebook($id: ID!) {
        node(id: $id) {
            ... on Notebook {
                ...NotebookFields
            }
        }
    }
    ${notebooksFragment}
`

export function fetchNotebook(id: Scalars['ID']): Observable<NotebookFields> {
    return requestGraphQL<FetchNotebookResult, FetchNotebookVariables>(fetchNotebookQuery, { id }).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (data.node?.__typename !== 'Notebook') {
                throw new Error('Not a valid notebook')
            }
            return data.node
        })
    )
}

const createNotebookMutation = gql`
    mutation CreateNotebook($notebook: NotebookInput!) {
        createNotebook(notebook: $notebook) {
            ...NotebookFields
        }
    }
    ${notebooksFragment}
`

export function createNotebook(variables: CreateNotebookVariables): Observable<NotebookFields> {
    return requestGraphQL<CreateNotebookResult, CreateNotebookVariables>(createNotebookMutation, variables).pipe(
        map(dataOrThrowErrors),
        map(data => data.createNotebook)
    )
}

const updateNotebookMutation = gql`
    mutation UpdateNotebook($id: ID!, $notebook: NotebookInput!) {
        updateNotebook(id: $id, notebook: $notebook) {
            ...NotebookFields
        }
    }
    ${notebooksFragment}
`

export function updateNotebook(variables: UpdateNotebookVariables): Observable<NotebookFields> {
    return requestGraphQL<UpdateNotebookResult, UpdateNotebookVariables>(updateNotebookMutation, variables).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateNotebook)
    )
}

const deleteNotebookMutation = gql`
    mutation DeleteNotebook($id: ID!) {
        deleteNotebook(id: $id) {
            alwaysNil
        }
    }
`

export function deleteNotebook(id: GQL.ID): Observable<DeleteNotebookResult> {
    return requestGraphQL<DeleteNotebookResult, DeleteNotebookVariables>(deleteNotebookMutation, { id }).pipe(
        map(dataOrThrowErrors)
    )
}
