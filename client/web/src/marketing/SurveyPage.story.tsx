import React from 'react'

import { DecoratorFn, Meta, Story } from '@storybook/react'

import { MockedStoryProvider } from '@sourcegraph/storybook/src/apollo/MockedStoryProvider'
import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { SurveyPage } from './SurveyPage'
import { submitSurveyMock } from './SurveyPage.mocks'
import { SurveyToast } from './SurveyToast'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/Survey',
    decorators: [decorator],
}

export default config

export const Page: Story = () => (
    <MockedStoryProvider mocks={[submitSurveyMock]}>
        <SurveyPage authenticatedUser={null} forceScore="10" />
    </MockedStoryProvider>
)

export const Toast: Story = () => <SurveyToast forceVisible={true} />
