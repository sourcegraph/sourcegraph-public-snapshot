import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'

import { queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    CreateSavedSearchResult,
    CreateSavedSearchVariables,
    DeleteSavedSearchResult,
    DeleteSavedSearchVariables,
    UpdateSavedSearchResult,
    UpdateSavedSearchVariables,
    Scalars,
} from '../graphql-operations'

export function fetchReposByQuery(query: string): Observable<{ name: string; url: string }[]> {
    return queryGraphQL(
        gql`
            query ReposByQuery($query: String!) {
                search(query: $query) {
                    results {
                        repositories {
                            name
                            url
                        }
                    }
                }
            }
        `,
        { query }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.results || !data.search.results.repositories) {
                throw createAggregateError(errors)
            }
            return data.search.results.repositories
        })
    )
}

const savedSearchFragment = gql`
    fragment SavedSearchFields on SavedSearch {
        id
        description
        notify
        notifySlack
        query
        namespace {
            __typename
            id
            namespaceName
        }
        slackWebhookURL
    }
`

export function fetchSavedSearches(): Observable<GQL.ISavedSearch[]> {
    return queryGraphQL(gql`
        query savedSearches {
            savedSearches {
                ...SavedSearchFields
            }
        }
        ${savedSearchFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.savedSearches) {
                throw createAggregateError(errors)
            }
            return data.savedSearches
        })
    )
}

export function fetchSavedSearch(id: Scalars['ID']): Observable<GQL.ISavedSearch> {
    return queryGraphQL(
        gql`
            query SavedSearch($id: ID!) {
                node(id: $id) {
                    ... on SavedSearch {
                        id
                        description
                        query
                        notify
                        notifySlack
                        slackWebhookURL
                        namespace {
                            id
                        }
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node as GQL.ISavedSearch)
    )
}

export function createSavedSearch(
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    userId: Scalars['ID'] | null,
    orgId: Scalars['ID'] | null
): Observable<void> {
    return requestGraphQL<CreateSavedSearchResult, CreateSavedSearchVariables>(
        gql`
            mutation CreateSavedSearch(
                $description: String!
                $query: String!
                $notifyOwner: Boolean!
                $notifySlack: Boolean!
                $userID: ID
                $orgID: ID
            ) {
                createSavedSearch(
                    description: $description
                    query: $query
                    notifyOwner: $notifyOwner
                    notifySlack: $notifySlack
                    userID: $userID
                    orgID: $orgID
                ) {
                    ...SavedSearchFields
                }
            }
            ${savedSearchFragment}
        `,
        {
            description,
            query,
            notifyOwner: notify,
            notifySlack,
            userID: userId,
            orgID: orgId,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export function updateSavedSearch(
    id: Scalars['ID'],
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    userId: Scalars['ID'] | null,
    orgId: Scalars['ID'] | null
): Observable<void> {
    return requestGraphQL<UpdateSavedSearchResult, UpdateSavedSearchVariables>(
        gql`
            mutation UpdateSavedSearch(
                $id: ID!
                $description: String!
                $query: String!
                $notifyOwner: Boolean!
                $notifySlack: Boolean!
                $userID: ID
                $orgID: ID
            ) {
                updateSavedSearch(
                    id: $id
                    description: $description
                    query: $query
                    notifyOwner: $notifyOwner
                    notifySlack: $notifySlack
                    userID: $userID
                    orgID: $orgID
                ) {
                    ...SavedSearchFields
                }
            }
            ${savedSearchFragment}
        `,
        {
            id,
            description,
            query,
            notifyOwner: notify,
            notifySlack,
            userID: userId,
            orgID: orgId,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export function deleteSavedSearch(id: Scalars['ID']): Observable<void> {
    return requestGraphQL<DeleteSavedSearchResult, DeleteSavedSearchVariables>(
        gql`
            mutation DeleteSavedSearch($id: ID!) {
                deleteSavedSearch(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}
