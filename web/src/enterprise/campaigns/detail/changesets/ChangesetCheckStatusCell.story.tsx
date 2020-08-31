import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'
import { ChangesetCheckState } from '../../../../graphql-operations'
import { capitalize } from 'lodash'
import { WebStory } from '../../../../components/WebStory'

const { add } = storiesOf('web/campaigns/ChangesetCheckStatusCell', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

for (const state of Object.values(ChangesetCheckState)) {
    add(capitalize(state), () => (
        <WebStory webStyles={webStyles}>{props => <ChangesetCheckStatusCell {...props} checkState={state} />}</WebStory>
    ))
}
