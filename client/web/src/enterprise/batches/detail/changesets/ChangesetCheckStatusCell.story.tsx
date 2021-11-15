import { storiesOf } from '@storybook/react'
import { capitalize } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetCheckState } from '../../../../graphql-operations'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'

const { add } = storiesOf('web/batches/ChangesetCheckStatusCell', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const state of Object.values(ChangesetCheckState)) {
    add(capitalize(state), () => (
        <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={state} />}</WebStory>
    ))
}
