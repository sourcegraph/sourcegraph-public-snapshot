import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { SurveyToast } from '.'

const decorator: DecoratorFn = story => <WebStory>{() => <div className="container mt-3">{story()}</div>}</WebStory>

const config: Meta = {
    title: 'web/SurveyToast',
    decorators: [decorator],
}

export default config

export const Toast: Story = () => <SurveyToast forceVisible={true} authenticatedUser={null} />
