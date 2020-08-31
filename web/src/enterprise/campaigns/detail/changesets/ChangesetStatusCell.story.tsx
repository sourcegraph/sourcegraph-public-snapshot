import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../../../components/WebStory'
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

add('Unpublished', () => <EnterpriseWebStory>{() => <ChangesetStatusUnpublished />}</EnterpriseWebStory>)
add('Closed', () => <EnterpriseWebStory>{() => <ChangesetStatusClosed />}</EnterpriseWebStory>)
add('Merged', () => <EnterpriseWebStory>{() => <ChangesetStatusMerged />}</EnterpriseWebStory>)
add('Open', () => <EnterpriseWebStory>{() => <ChangesetStatusOpen />}</EnterpriseWebStory>)
add('Deleted', () => <EnterpriseWebStory>{() => <ChangesetStatusDeleted />}</EnterpriseWebStory>)
add('Error', () => <EnterpriseWebStory>{() => <ChangesetStatusError />}</EnterpriseWebStory>)
add('Processing', () => <EnterpriseWebStory>{() => <ChangesetStatusProcessing />}</EnterpriseWebStory>)
