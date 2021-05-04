import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { CreationIntroPage } from './CreationIntroPage';

const { add } = storiesOf('web/insights/CreationInsightIntroPage', module)
    .addDecorator(story => <WebStory>{() => story()}</WebStory>)
    .addParameters({
        chromatic: {
            viewports: [320, 576, 978, 1440],
        },
    })

add('Page', () => (
    <CreationIntroPage />
))
