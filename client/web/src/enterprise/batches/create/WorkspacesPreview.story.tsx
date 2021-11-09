import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { WorkspacesPreview } from './WorkspacesPreview'

const { add } = storiesOf('web/batches/CreateBatchChangePage/WorkspacesPreview', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('empty', () => <WebStory>{props => <WorkspacesPreview {...props} />}</WebStory>)
