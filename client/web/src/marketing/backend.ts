import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'

import { requestGraphQL } from '../backend/graphql'
import type {
    FetchAllUsersWithSurveyResponsesResult,
    FetchAllUsersWithSurveyResponsesVariables,
    FetchSurveyResponseAggregatesResult,
    FetchSurveyResponseAggregatesVariables,
    FetchSurveyResponsesResult,
    FetchSurveyResponsesVariables,
    UserActivePeriod,
} from '../graphql-operations'

/**
 * Fetches survey responses.
 */
export function fetchAllSurveyResponses(): Observable<FetchSurveyResponsesResult['surveyResponses']> {
    return requestGraphQL<FetchSurveyResponsesResult, FetchSurveyResponsesVariables>(
        gql`
            query FetchSurveyResponses {
                surveyResponses {
                    nodes {
                        ...SurveyResponseFields
                    }
                    totalCount
                }
            }

            fragment SurveyResponseFields on SurveyResponse {
                user {
                    id
                    username
                    emails {
                        email
                    }
                }
                email
                score
                reason
                better
                otherUseCase
                createdAt
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.surveyResponses)
    )
}

/**
 * Fetches users, with their survey responses.
 */
export function fetchAllUsersWithSurveyResponses(args: {
    activePeriod?: UserActivePeriod
    first?: number
    query?: string
}): Observable<FetchAllUsersWithSurveyResponsesResult['users']> {
    return requestGraphQL<FetchAllUsersWithSurveyResponsesResult, FetchAllUsersWithSurveyResponsesVariables>(
        gql`
            query FetchAllUsersWithSurveyResponses($activePeriod: UserActivePeriod, $first: Int, $query: String) {
                users(activePeriod: $activePeriod, first: $first, query: $query) {
                    nodes {
                        ...UserWithSurveyResponseFields
                    }
                    totalCount
                }
            }

            fragment UserWithSurveyResponseFields on User {
                id
                username
                emails {
                    email
                }
                surveyResponses {
                    score
                    reason
                    better
                    otherUseCase
                    createdAt
                }
                usageStatistics {
                    lastActiveTime
                }
            }
        `,
        { activePeriod: args.activePeriod ?? null, first: args.first ?? null, query: args.query ?? null }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}

/**
 * Fetches survey response aggregate data.
 */
export function fetchSurveyResponseAggregates(): Observable<FetchSurveyResponseAggregatesResult['surveyResponses']> {
    return requestGraphQL<FetchSurveyResponseAggregatesResult, FetchSurveyResponseAggregatesVariables>(
        gql`
            query FetchSurveyResponseAggregates {
                surveyResponses {
                    ...SurveyResponseAggregateFields
                }
            }

            fragment SurveyResponseAggregateFields on SurveyResponseConnection {
                totalCount
                last30DaysCount
                averageScore
                netPromoterScore
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.surveyResponses)
    )
}
