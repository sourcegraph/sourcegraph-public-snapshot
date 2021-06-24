import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchChangesChangelogAlert } from './BatchChangesChangelogAlert'

const { add } = storiesOf('web/batches/BatchChangesChangelogAlert', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Changelog', () => <EnterpriseWebStory>{() => <BatchChangesChangelogAlert />}</EnterpriseWebStory>)
