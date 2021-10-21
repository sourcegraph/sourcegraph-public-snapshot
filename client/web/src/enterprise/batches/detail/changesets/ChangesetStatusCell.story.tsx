import { storiesOf } from '@storybook/react'
import { capitalize } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetState } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'

const { add } = storiesOf('web/batches/ChangesetStatusCell', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const state of Object.values(ChangesetState)) {
    add(capitalize(state), () => (
        <WebStory>{() => <ChangesetStatusCell state={state} className="d-flex text-muted" />}</WebStory>
    ))
}
