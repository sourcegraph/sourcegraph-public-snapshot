import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { ClosedNotice } from './ClosedNotice'

const { add } = storiesOf('web/batches/details/ClosedNotice', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Batch change closed', () => <WebStory>{() => <ClosedNotice closedAt="2021-02-02" />}</WebStory>)
