import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import type { SubmitSurveyResult, SubmitSurveyVariables } from '../../graphql-operations'

import { SUBMIT_SURVEY } from './SurveyForm'

export const mockVariables: SubmitSurveyVariables['input'] = {
    score: 10,
    otherUseCase: 'Learn best practices',
    better: 'Add this new feature',
    email: '',
}

export const submitSurveyMock: MockedResponse<SubmitSurveyResult> = {
    request: {
        query: getDocumentNode(SUBMIT_SURVEY),
        variables: {
            input: mockVariables,
        },
    },
    result: {
        data: {
            submitSurvey: {
                alwaysNil: null,
                __typename: 'EmptyResponse',
            },
        },
    },
}
