import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'

const { add } = storiesOf('web/batches/BatchChangesChangelogAlert', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Changelog', () => <WebStory>{() => <BatchChangesChangelogAlert />}</WebStory>)
