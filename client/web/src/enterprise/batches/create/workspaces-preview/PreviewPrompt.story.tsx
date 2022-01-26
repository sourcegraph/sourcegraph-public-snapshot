import { boolean } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '@sourcegraph/web/src/components/WebStory'

import { PreviewPrompt } from './PreviewPrompt'

const { add } = storiesOf(
    'web/batches/CreateBatchChangePage/WorkspacesPreview/PreviewPrompt',
    module
).addDecorator(story => <div className="p-3 container d-flex flex-column align-items-center">{story()}</div>)

add('initial', () => (
    <WebStory>
        {props => <PreviewPrompt {...props} form="Initial" disabled={boolean('Disabled', false)} preview={noop} />}
    </WebStory>
))

add('error', () => (
    <WebStory>
        {props => <PreviewPrompt {...props} form="Error" disabled={boolean('Disabled', false)} preview={noop} />}
    </WebStory>
))

add('update', () => (
    <WebStory>
        {props => <PreviewPrompt {...props} form="Update" disabled={boolean('Disabled', false)} preview={noop} />}
    </WebStory>
))
