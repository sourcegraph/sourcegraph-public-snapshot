import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'

import { Description } from './Description'

const { add } = storiesOf('web/batches/Description', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Overview', () => (
    <WebStory>
        {props => (
            <Description
                {...props}
                description="This is an awesome batch change. It will do great things to your codebase."
            />
        )}
    </WebStory>
))
