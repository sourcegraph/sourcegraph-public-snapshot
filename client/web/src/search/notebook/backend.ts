import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'

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
    return requestGraphQL<ListNotebooksResult, ListNotebooksVariables>(
        gql`
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
        `,
        {
            first,
            after: after ?? null,
            query: query ?? null,
            creatorUserID: creatorUserID ?? null,
            orderBy: orderBy ?? GQL.NotebooksOrderBy.NOTEBOOK_UPDATED_AT,
            descending: descending ?? false,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.notebooks)
    )
}

export function fetchNotebook(id: Scalars['ID']): Observable<NotebookFields> {
    return requestGraphQL<FetchNotebookResult, FetchNotebookVariables>(
        gql`
            query FetchNotebook($id: ID!) {
                node(id: $id) {
                    ... on Notebook {
                        ...NotebookFields
                    }
                }
            }
            ${notebooksFragment}
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (data.node?.__typename !== 'Notebook') {
                throw new Error('Not a valid notebook')
            }
            return data.node
        })
    )
}

export function createNotebook(variables: CreateNotebookVariables): Observable<NotebookFields> {
    return requestGraphQL<CreateNotebookResult, CreateNotebookVariables>(
        gql`
            mutation CreateNotebook($notebook: NotebookInput!) {
                createNotebook(notebook: $notebook) {
                    ...NotebookFields
                }
            }
            ${notebooksFragment}
        `,
        variables
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.createNotebook)
    )
}

export function updateNotebook(variables: UpdateNotebookVariables): Observable<NotebookFields> {
    return requestGraphQL<UpdateNotebookResult, UpdateNotebookVariables>(
        gql`
            mutation UpdateNotebook($id: ID!, $notebook: NotebookInput!) {
                updateNotebook(id: $id, notebook: $notebook) {
                    ...NotebookFields
                }
            }
            ${notebooksFragment}
        `,
        variables
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateNotebook)
    )
}

export function deleteNotebook(id: GQL.ID): Observable<DeleteNotebookResult> {
    return requestGraphQL<DeleteNotebookResult, DeleteNotebookVariables>(
        gql`
            mutation DeleteNotebook($id: ID!) {
                deleteNotebook(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(map(dataOrThrowErrors))
}
