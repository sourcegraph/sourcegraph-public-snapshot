import { MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { getDocumentNode } from '@sourcegraph/http-client'
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
        /**
         * Uncomment this once Storybook is upgraded to v6.4.* and the `play` function
         * is used to show the feedback prompt component.
         *
         * https://www.chromatic.com/docs/hoverfocus#javascript-triggered-hover-states
         */
        // chromatic: { disableSnapshot: false },
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
}

export default config

export const FeedbackPromptExample: Story = () => (
    <>
        <h1>This is a feedbackPrompt</h1>
        <FeedbackPrompt routes={[]} />
    </>
)
