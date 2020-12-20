import { Observable } from 'rxjs'
import { map, mapTo } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import { requestGraphQL } from '../backend/graphql'
import { SurveyResponse } from './SurveyPage'
import {
    FetchAllUsersWithSurveyResponsesResult,
    FetchAllUsersWithSurveyResponsesVariables,
    FetchSurveyResponseAggregatesResult,
    FetchSurveyResponseAggregatesVariables,
    FetchSurveyResponsesResult,
    FetchSurveyResponsesVariables,
    RequestTrialResult,
    RequestTrialVariables,
    SubmitSurveyResult,
    SubmitSurveyVariables,
    UserActivePeriod,
} from '../graphql-operations'

/**
 * Submits a user satisfaction survey.
 */
export function submitSurvey(input: SurveyResponse): Observable<void> {
    return requestGraphQL<SubmitSurveyResult, SubmitSurveyVariables>(
        gql`
            mutation SubmitSurvey($input: SurveySubmissionInput!) {
                submitSurvey(input: $input) {
                    alwaysNil
                }
            }
        `,
        { input }
    ).pipe(map(dataOrThrowErrors), mapTo(undefined))
}

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

/**
 * Submits a request for a Sourcegraph Enterprise trial license.
 */
export const submitTrialRequest = (email: string): void => {
    requestGraphQL<RequestTrialResult, RequestTrialVariables>(
        gql`
            mutation RequestTrial($email: String!) {
                requestTrial(email: $email) {
                    alwaysNil
                }
            }
        `,
        { email }
    )
        .pipe(map(dataOrThrowErrors), mapTo(undefined))
        // eslint-disable-next-line rxjs/no-ignored-subscription
        .subscribe({
            error: () => {
                // Swallow errors since the form submission is a non-blocking request that happens during site-init
                // if a trial license key is requested.
            },
        })
}
