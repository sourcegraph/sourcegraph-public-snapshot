import { MockedResponse } from '@apollo/client/testing'
import { Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { getDocumentNode } from '@sourcegraph/shared/src/graphql/graphql'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { FeedbackPrompt, SUBMIT_HAPPINESS_FEEDBACK_QUERY } from '.'

const mockRequest: MockedResponse = {
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

const config: Meta = {
    title: 'wildcard/FeedbackPrompt',

    decorators: [
        story => (
            <BrandedStory mocks={[mockRequest]} styles={webStyles}>
                {() => <div className="container mt-3">{story()}</div>}
            </BrandedStory>
        ),
    ],
    parameters: {
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
}

export default config

export const FeedbackPromptExample = () => (
    <>
        <h1>This is a feedbackPrompt</h1>
        <FeedbackPrompt routes={[]} />
    </>
)
