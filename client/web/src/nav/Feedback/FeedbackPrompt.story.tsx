import { storiesOf } from '@storybook/react'
import React from 'react'
import { createBrowserHistory } from 'history'
import { WebStory } from '../../components/WebStory'
import { FeedbackPrompt } from './FeedbackPrompt'

const history = createBrowserHistory()

const { add } = storiesOf('web/nav', module)

add('Feedback Widget', () => <WebStory>{() => <FeedbackPrompt open={true} history={history} />}</WebStory>, {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/9FprSCL6roIZotcWMvJZuE/Improving-user-feedback-channels?node-id=300%3A3555',
    },
})
