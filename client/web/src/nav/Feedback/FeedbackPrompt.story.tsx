import { MockedResponse } from '@apollo/client/testing'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'

import { WebStory } from '../../components/WebStory'
import { SubmitHappinessFeedbackResult } from '../../graphql-operations'

import { FeedbackPrompt, SUBMIT_HAPPINESS_FEEDBACK_QUERY } from './FeedbackPrompt'

const { add } = storiesOf('web/nav', module)

const mockRequest: MockedResponse<SubmitHappinessFeedbackResult> = {
    request: {
        query: getDocumentNode(SUBMIT_HAPPINESS_FEEDBACK_QUERY),
    },
    result: {
        data: {
            submitHappinessFeedback: {
                alwaysNil: null,
            },
        },
    },
}

add(
    'Feedback Widget',
    () => <WebStory mocks={[mockRequest]}>{() => <FeedbackPrompt open={true} routes={[]} />}</WebStory>,
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/9FprSCL6roIZotcWMvJZuE/Improving-user-feedback-channels?node-id=300%3A3555',
        },
    }
)
