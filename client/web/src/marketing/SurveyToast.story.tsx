import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { SurveyToast } from './SurveyToast'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/SurveyToast',
    decorators: [decorator],
}

export default config

export const Toast: Story = () => <SurveyToast forceVisible={true} />
