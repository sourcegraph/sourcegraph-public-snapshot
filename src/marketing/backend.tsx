import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { SurveyResponse } from './SurveyPage'

/**
 * Submits a user satisfaction survey.
 */
export function submitSurvey(input: SurveyResponse): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SubmitSurvey($input: SurveySubmissionInput!) {
                submitSurvey(input: $input) {
                    alwaysNil
                }
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

/**
 * Fetches survey responses.
 */
export function fetchAllSurveyResponses(args: { first?: number }): Observable<GQL.ISurveyResponseConnection> {
    return queryGraphQL(
        gql`
            query FetchSurveyResponses() {
                surveyResponses {
                    nodes {
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
                    totalCount
                }
            }
        `,
        { ...args }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.surveyResponses)
    )
}

/**
 * Fetches users, with their survey responses.
 */
export function fetchAllUsersWithSurveyResponses(args: {
    activePeriod?: GQL.UserActivePeriod
    first?: number
    query?: string
}): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query FetchAllUsersWithSurveyResponses($activePeriod: UserActivePeriod, $first: Int, $query: String) {
                users(activePeriod: $activePeriod, first: $first, query: $query) {
                    nodes {
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
                        activity {
                            lastActiveTime
                        }
                    }
                    totalCount
                }
            }
        `,
        { ...args }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}

export type SurveyResponseConnectionAggregates = Exclude<GQL.ISurveyResponseConnection, 'nodes'>

/**
 * Fetches survey response aggregate data.
 */
export function fetchSurveyResponseAggregates(): Observable<SurveyResponseConnectionAggregates> {
    return queryGraphQL(
        gql`
            query FetchSurveyResponseAggregates() {
                surveyResponses {
                    totalCount
                    last30DaysCount
                    averageScore
                    netPromoterScore
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.surveyResponses)
    )
}
