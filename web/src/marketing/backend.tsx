import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { createAggregateError } from '../util/errors'
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
        map(({ data, errors }) => {
            if (errors) {
                throw createAggregateError(errors)
            }
            return
        })
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
                }
            }
        `,
        { ...args }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.surveyResponses || !data.surveyResponses.nodes) {
                throw createAggregateError(errors)
            }
            return data.surveyResponses
        })
    )
}
