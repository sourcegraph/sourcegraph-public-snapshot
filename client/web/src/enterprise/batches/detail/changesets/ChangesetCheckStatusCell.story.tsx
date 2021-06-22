import { storiesOf } from '@storybook/react'
import { capitalize } from 'lodash'
import React from 'react'

import { ChangesetCheckState } from '../../../../graphql-operations'
import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'

const { add } = storiesOf('web/batches/ChangesetCheckStatusCell', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const state of Object.values(ChangesetCheckState)) {
    add(capitalize(state), () => (
        <EnterpriseWebStory>{props => <ChangesetCheckStatusCell {...props} checkState={state} />}</EnterpriseWebStory>
    ))
}
