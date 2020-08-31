import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { HiddenChangesetSpecNode } from './HiddenChangesetSpecNode'
import { addDays } from 'date-fns'
import { HiddenChangesetSpecFields, ChangesetSpecType } from '../../../graphql-operations'
import { WebStory } from '../../../components/WebStory'

const { add } = storiesOf('web/campaigns/apply/HiddenChangesetSpecNode', module).addDecorator(story => (
    <div className="p-3 container web-content changeset-spec-list__grid">{story()}</div>
))

export const hiddenChangesetSpecStories: Record<string, HiddenChangesetSpecFields> = {
    'Import changeset': {
        __typename: 'HiddenChangesetSpec',
        id: 'someidh1',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
    },
    'Create changeset': {
        __typename: 'HiddenChangesetSpec',
        id: 'someidh2',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.BRANCH,
    },
}

for (const storyName of Object.keys(hiddenChangesetSpecStories)) {
    add(storyName, () => (
        <WebStory webStyles={webStyles}>
            {props => <HiddenChangesetSpecNode {...props} node={hiddenChangesetSpecStories[storyName]} />}
        </WebStory>
    ))
}
