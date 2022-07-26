import { Observable, of } from 'rxjs'
import { map, tap } from 'rxjs/operators'

import { createAggregateError } from '@sourcegraph/common'
import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import * as GQL from '@sourcegraph/shared/src/schema'

import { InvitableCollaborator } from '../auth/welcome/InviteCollaborators/InviteCollaborators'
import { queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    EventLogsDataResult,
    EventLogsDataVariables,
    CreateSavedSearchResult,
    CreateSavedSearchVariables,
    DeleteSavedSearchResult,
    DeleteSavedSearchVariables,
    UpdateSavedSearchResult,
    UpdateSavedSearchVariables,
    Scalars,
    InvitableCollaboratorsResult,
    InvitableCollaboratorsVariables,
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

export interface EventLogResult {
    totalCount: number
    nodes: { argument: string | null; timestamp: string; url: string }[]
    pageInfo: { hasNextPage: boolean }
}

function fetchEvents(userId: Scalars['ID'], first: number, eventName: string): Observable<EventLogResult | null> {
    if (!userId) {
        return of(null)
    }

    const result = requestGraphQL<EventLogsDataResult, EventLogsDataVariables>(
        gql`
            query EventLogsData($userId: ID!, $first: Int, $eventName: String!) {
                node(id: $userId) {
                    ... on User {
                        __typename
                        eventLogs(first: $first, eventName: $eventName) {
                            nodes {
                                argument
                                timestamp
                                url
                            }
                            pageInfo {
                                hasNextPage
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        { userId, first: first ?? null, eventName }
    )

    return result.pipe(
        map(dataOrThrowErrors),
        map(
            (data: EventLogsDataResult): EventLogResult => {
                if (!data.node || data.node.__typename !== 'User') {
                    throw new Error('User not found')
                }
                return data.node.eventLogs
            }
        )
    )
}

export function fetchRecentSearches(userId: Scalars['ID'], first: number): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'SearchResultsQueried')
}

export function fetchRecentFileViews(userId: Scalars['ID'], first: number): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'ViewBlob')
}

export function fetchCollaborators(userId: Scalars['ID']): Observable<InvitableCollaborator[]> {
    if (!userId) {
        return of([])
    }

    const result = requestGraphQL<InvitableCollaboratorsResult, InvitableCollaboratorsVariables>(
        gql`
            query InvitableCollaborators {
                currentUser {
                    invitableCollaborators {
                        name
                        email
                        displayName
                        avatarURL
                    }
                }
            }
        `,
        {}
    )

    return result.pipe(
        map(dataOrThrowErrors),
        map(
            (data: InvitableCollaboratorsResult): InvitableCollaborator[] =>
                data.currentUser?.invitableCollaborators ?? []
        )
    )
}

export function fetchSearchResults(query: string, patternType: string): string {
    console.log('fetchSearchResults', query, patternType)
    const result = requestGraphQL<any, any>(
        gql`
            query SearchResults($query: String!, $patternType: SearchPatternType) {
                search(query: $query, patternType: $patternType) {
                    results {
                        results {
                            __typename
                            ... on CommitSearchResult {
                                url
                                commit {
                                    subject
                                    author {
                                        date
                                        person {
                                            displayName
                                        }
                                    }
                                }
                            }
                            ... on Repository {
                                name
                                externalURLs {
                                    url
                                }
                            }
                            ... on FileMatch {
                                repository {
                                    name
                                    externalURLs {
                                        url
                                    }
                                }
                                file {
                                    path
                                    canonicalURL
                                    externalURLs {
                                        url
                                    }
                                }
                                lineMatches {
                                    preview
                                    offsetAndLengths
                                }
                            }
                        }
                    }
                }
            }
        `,
        { query, patternType }
    ).pipe(tap(console.log))

    return result

    // const results = result.search.results.results

    // if (!results?.length || !results[0]) {
    //     throw new Error('No results to be exported.')
    // }
    // const headers =
    //     results[0].__typename !== 'CommitSearchResult'
    //         ? [
    //               'Match type',
    //               'Repository',
    //               'Repository external URL',
    //               'File path',
    //               'File URL',
    //               'File external URL',
    //               'Search matches',
    //           ]
    //         : ['Date', 'Author', 'Subject', 'Commit URL']
    // const csvData = [
    //     headers,
    //     ...results.map(r => {
    //         switch (r.__typename) {
    //             // on FileMatch
    //             case 'FileMatch':
    //                 const searchMatches = r.lineMatches
    //                     .map(line =>
    //                         line.offsetAndLengths
    //                             .map(offset => line.preview?.substring(offset[0], offset[0] + offset[1]))
    //                             .join(' ')
    //                     )
    //                     .join(' ')

    //                 return [
    //                     r.__typename,
    //                     r.repository.name,
    //                     r.repository.externalURLs[0]?.url,
    //                     r.file.path,
    //                     new URL(r.file.canonicalURL, sourcegraph.internal.sourcegraphURL).toString(),
    //                     r.file.externalURLs[0]?.url,
    //                     truncateMatches(searchMatches),
    //                 ].map(s => JSON.stringify(s))
    //             // on Repository
    //             case 'Repository':
    //                 return [r.__typename, r.name, r.externalURLs[0]?.url].map(s => JSON.stringify(s))
    //             // on CommitSearchResult
    //             case 'CommitSearchResult':
    //                 return [r.commit.author.date, r.commit.author.person.displayName, r.commit.subject, r.url].map(s =>
    //                     JSON.stringify(s)
    //                 )
    //             // If no typename can be found
    //             default:
    //                 throw new Error('Please try another query.')
    //         }
    //     }),
    // ]

    // const encodedData = encodeURIComponent(csvData.map(row => row.join(',')).join('\n'))

    // return encodedData
}
