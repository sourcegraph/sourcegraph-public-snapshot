import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../../../components/WebStory'
import webStyles from '../../../../enterprise.scss'
import {
    ChangesetStatusUnpublished,
    ChangesetStatusProcessing,
    ChangesetStatusClosed,
    ChangesetStatusMerged,
    ChangesetStatusOpen,
    ChangesetStatusDeleted,
    ChangesetStatusError,
} from './ChangesetStatusCell'

const { add } = storiesOf('web/campaigns/ChangesetStatusCell', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Unpublished', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusUnpublished />}</WebStory>)
add('Closed', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusClosed />}</WebStory>)
add('Merged', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusMerged />}</WebStory>)
add('Open', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusOpen />}</WebStory>)
add('Deleted', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusDeleted />}</WebStory>)
add('Error', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusError />}</WebStory>)
add('Processing', () => <WebStory webStyles={webStyles}>{() => <ChangesetStatusProcessing />}</WebStory>)
